package ecs_provisioner

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/models"
	"github.com/rafaeleyng/pushaas/pushaas/provisioners"
)

type (
	ecsProvisionerConfig struct {
		imagePushApi     string
		imagePushAgent   string
		imagePushStream  string
		region           string
		cluster          string
		logsStreamPrefix string
		logsGroup        string
		securityGroup    string
		subnet           string
		dnsNamespace     string
	}

	ecsProvisioner struct {
		logger                *zap.Logger
		session               *session.Session
		provisionerConfig     *ecsProvisionerConfig
		pushRedisProvisioner  EcsPushRedisProvisioner
		pushStreamProvisioner EcsPushStreamProvisioner
		pushApiProvisioner    EcsPushApiProvisioner
		attempts              int
		interval              time.Duration
	}
)

/////////////////////////////////////////////////////////////////////////////////
//// service discovery
/////////////////////////////////////////////////////////////////////////////////
//func listServiceDiscoveryServices(svc *servicediscovery.ServiceDiscovery) *servicediscovery.ListServicesOutput {
//	input := &servicediscovery.ListServicesInput{}
//
//	output, err := svc.ListServices(input)
//	if err != nil {
//		fmt.Println("========== FAILED ListServices discovery ==========")
//	}
//	fmt.Println("========== ListServices discovery ==========")
//	fmt.Println(output.GoString())
//	return output
//}
//
//func deleteServiceDiscoveryServices(svc *servicediscovery.ServiceDiscovery) {
//	discoveryServices := listServiceDiscoveryServices(svc)
//	services := discoveryServices.Services
//	for _, s := range services {
//		if strings.HasPrefix(*s.Name, "push-") {
//			deleteInput := &servicediscovery.DeleteServiceInput{Id:s.Id}
//			deleteOutput, err := svc.DeleteService(deleteInput)
//			if err != nil {
//				fmt.Println("========== FAILED DeleteService discovery ==========", *s.Name)
//			}
//			fmt.Println("========== DeleteService discovery ==========")
//			fmt.Println(deleteOutput.GoString())
//		} else {
//			fmt.Println("do not delete", s.Name)
//		}
//	}
//}

//func (p *ecsProvisioner) Test() {
//	//instance := &models.Instance{
//	//	Name: "instance-33",
//	//}
//}

type (
	provisionPushRedisResult struct {
		serviceDiscovery *servicediscovery.CreateServiceOutput
		service          *ecs.CreateServiceOutput
	}

	provisionPushApiResult struct {
		serviceDiscovery *servicediscovery.CreateServiceOutput
		taskDefinition   *ecs.RegisterTaskDefinitionOutput
		service          *ecs.CreateServiceOutput
	}

	provisionPushStreamResult struct {
		serviceDiscovery *servicediscovery.CreateServiceOutput
		taskDefinition   *ecs.RegisterTaskDefinitionOutput
		service          *ecs.CreateServiceOutput
	}
)

func (p *ecsProvisioner) TODO() {

	// TODO mover para provisioner
	//eni, err := DescribePushStreamTaskNetworkInterface(instance, ecsSvc, ec2Svc, provisionerConfig)
	//if err != nil {
	//	return nil, errors.New("failed to obtain push-stream public IP to create push-api task definition")
	//}

}

func (p *ecsProvisioner) waitSuccess(ch chan bool, evaluationFn func() bool) {
	for i := 0; i < p.attempts; i++ {
		isLastAttempt := i + 1 == p.attempts
		isSuccess := evaluationFn()

		if !isSuccess {
			if isLastAttempt {
				ch <- false
				break
			}
			time.Sleep(p.interval)
			continue
		}
		ch <- true
	}
}

func (p *ecsProvisioner) waitService(instance *models.Instance, ecsSvc *ecs.ECS, ch chan bool, describeServiceFunc func(*models.Instance, *ecs.ECS, *ecsProvisionerConfig) (*ecs.DescribeServicesOutput, error)) {
	p.waitSuccess(ch, func() bool {
		serviceResult, err := describeServiceFunc(instance, ecsSvc, p.provisionerConfig)
		failed := err != nil || len(serviceResult.Services) == 0 || serviceResult.Services[0].RunningCount == aws.Int64(0)
		return !failed
	})
}

func (p *ecsProvisioner) waitTaskNetworkInterface(instance *models.Instance, ecsSvc *ecs.ECS, ec2Svc *ec2.EC2, ch chan bool, describeEniFunc func(instance *models.Instance, ecsSvc *ecs.ECS, ec2Svc *ec2.EC2, provisionerConfig *ecsProvisionerConfig) (*ec2.DescribeNetworkInterfacesOutput, error)) {
	p.waitSuccess(ch, func() bool {
		_, err := describeEniFunc(instance, ecsSvc, ec2Svc, p.provisionerConfig)
		failed := err != nil
		return !failed
	})
}

func (p *ecsProvisioner) Provision(instance *models.Instance) provisioners.PushServiceProvisionResult {
	p.logger.Info("starting provision for instance", zap.Any("instance", instance))

	iamSvc := iam.New(p.session)
	ecsSvc := ecs.New(p.session)
	ec2Svc := ec2.New(p.session)
	serviceDiscoverySvc := servicediscovery.New(p.session)

	var err error
	role, err := getIamRole(iamSvc)
	if err != nil {
		p.logger.Error("failed while provisioning instance, failed to get iam role", zap.Any("instance", instance), zap.Error(err))
		return provisioners.PushServiceProvisionResultFailure
	}

	/*
		push-redis
	 */
	chRedis := make(chan bool)
	resultPushRedis, err := p.pushRedisProvisioner.Provision(instance, ecsSvc, serviceDiscoverySvc, p.provisionerConfig)
	if err != nil {
		p.logger.Error("failed while provisioning instance, failed to provision push-redis", zap.Any("instance", instance), zap.Error(err))
		return provisioners.PushServiceProvisionResultFailure
	}

	/*
		push-stream
	*/
	chStream := make(chan bool)
	resultPushStream, err := p.pushStreamProvisioner.Provision(instance, ecsSvc, serviceDiscoverySvc, role, p.provisionerConfig)
	if err != nil {
		p.logger.Error("failed while provisioning instance, failed to provision push-stream", zap.Any("instance", instance))
		return provisioners.PushServiceProvisionResultFailure
	}

	go p.waitService(instance, ecsSvc, chRedis, p.pushRedisProvisioner.DescribeService)
	p.logger.Info("waiting for push-redis to become available", zap.Any("instance", instance))
	isRedisUp := <- chRedis
	if !isRedisUp {
		p.logger.Error("push-redis failed to start", zap.Any("instance", instance))
		// TODO deprovision
	}
	p.logger.Info("push-redis started", zap.Any("instance", instance))

	go p.waitService(instance, ecsSvc, chStream, p.pushStreamProvisioner.DescribeService)
	p.logger.Info("waiting for push-stream to become available", zap.Any("instance", instance))
	isStreamUp := <- chStream
	if !isStreamUp {
		p.logger.Error("push-stream failed to start", zap.Any("instance", instance))
		// TODO deprovision
	}
	p.logger.Info("push-stream started", zap.Any("instance", instance))

	someFailed := !isRedisUp || !isStreamUp
	if someFailed {
		return provisioners.PushServiceProvisionResultFailure
	}

	/*
		push-stream ENI
	*/
	chEni := make(chan bool)
	go p.waitTaskNetworkInterface(instance, ecsSvc, ec2Svc, chEni, p.pushStreamProvisioner.DescribePushStreamTaskNetworkInterface)
	p.logger.Info("waiting for push-stream ENI to become available so we can start provisioning push-api", zap.Any("instance", instance), zap.Error(err))
	isEniUp := <- chEni
	if !isEniUp {
		p.logger.Error("push-stream ENI failed to start", zap.Any("instance", instance))
		// TODO deprovision
	}
	p.logger.Info("push-stream ENI started", zap.Any("instance", instance))

	eni, err := p.pushStreamProvisioner.DescribePushStreamTaskNetworkInterface(instance, ecsSvc, ec2Svc, p.provisionerConfig)
	if err != nil {
		p.logger.Error("failed to retrieve push-stream ENI", zap.Any("instance", instance), zap.Error(err))
		return provisioners.PushServiceProvisionResultFailure
	}

	/*
		push-api
	*/
	chApi := make(chan bool)
	resultPushApi, err := p.pushApiProvisioner.Provision(instance, ecsSvc, serviceDiscoverySvc, ec2Svc, role, eni, p.provisionerConfig)
	if err != nil {
		p.logger.Error("failed while provisioning instance, failed to provision push-api", zap.Any("instance", instance), zap.Error(err))
		return provisioners.PushServiceProvisionResultFailure
	}

	go p.waitService(instance, ecsSvc, chApi, p.pushApiProvisioner.DescribeService)
	p.logger.Info("waiting for push-api to become available", zap.Any("instance", instance), zap.Error(err))
	isApiUp := <- chApi
	if !isApiUp {
		p.logger.Error("push-api failed to start", zap.Any("instance", instance))
		// TODO deprovision
	}
	p.logger.Info("push-api started", zap.Any("instance", instance))

	p.logger.Info(
		"finishing provision for instance",
		zap.Any("instance", instance),
		zap.Any("resultPushRedis", resultPushRedis),
		zap.Any("resultPushStream", resultPushStream),
		zap.Any("resultPushApi", resultPushApi),
	)

	return provisioners.PushServiceProvisionResultSuccess
}

func (p *ecsProvisioner) Deprovision(instance *models.Instance) provisioners.PushServiceDeprovisionResult {
	p.logger.Info("starting deprovision for instance", zap.Any("instance", instance))

	return provisioners.PushServiceDeprovisionResultSuccess
}

func NewEcsPushServiceProvisioner(
	config *viper.Viper,
	logger *zap.Logger,
	pushRedisProvisioner EcsPushRedisProvisioner,
	pushStreamProvisioner EcsPushStreamProvisioner,
	pushApiProvisioner EcsPushApiProvisioner,
) (provisioners.PushServiceProvisioner, error) {
	imagePushApi := config.GetString("provisioner.ecs.image-push-api")
	imagePushAgent := config.GetString("provisioner.ecs.image-push-agent")
	imagePushStream := config.GetString("provisioner.ecs.image-push-stream")

	region := config.GetString("provisioner.ecs.region")
	cluster := config.GetString("provisioner.ecs.cluster")
	logsStreamPrefix := config.GetString("provisioner.ecs.logs-stream-prefix")
	logsGroup := config.GetString("provisioner.ecs.logs-group")

	attempts := config.GetInt("provisioner.ecs.attempts")
	interval := config.GetDuration("provisioner.ecs.interval")

	// required vars
	// TODO comes from `scripts/40-pushaas/30-create-cluster/terraform.tfstate`, should create specific for each part of push service
	securityGroup := config.GetString("provisioner.ecs.security-group")
	// TODO comes from `scripts/10-vpc/10-create-vpc/terraform.tfstate` - should have public and private
	subnet := config.GetString("provisioner.ecs.subnet")
	// TODO comes from `scripts/30-dns/10-create-namespace/terraform.tfstate`
	dnsNamespace := config.GetString("provisioner.ecs.dns-namespace")

	requiredVars := []string{
		securityGroup,
		subnet,
		dnsNamespace,
	}

	for _, v := range requiredVars {
		if v == "" {
			return nil, errors.New(fmt.Sprintf("ecsProvisioner env required and not set: %s", v))
		}
	}

	return &ecsProvisioner{
		logger:  logger,
		session: session.Must(session.NewSession()),
		attempts: attempts,
		interval: interval,
		provisionerConfig: &ecsProvisionerConfig{
			imagePushApi:     imagePushApi,
			imagePushAgent:   imagePushAgent,
			imagePushStream:  imagePushStream,
			region:           region,
			cluster:          cluster,
			logsStreamPrefix: logsStreamPrefix,
			logsGroup:        logsGroup,
			securityGroup:    securityGroup,
			subnet:           subnet,
			dnsNamespace:     dnsNamespace,
		},
		pushRedisProvisioner:  pushRedisProvisioner,
		pushStreamProvisioner: pushStreamProvisioner,
		pushApiProvisioner:    pushApiProvisioner,
	}, nil
}
