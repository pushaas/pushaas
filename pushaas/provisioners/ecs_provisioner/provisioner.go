package ecs_provisioner

import (
	"errors"
	"fmt"

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
//		panic(err)
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
//				panic(err)
//			}
//			fmt.Println("========== DeleteService discovery ==========")
//			fmt.Println(deleteOutput.GoString())
//		} else {
//			fmt.Println("do not delete", s.Name)
//		}
//	}
//}

func (p *ecsProvisioner) Test() {
	//instance := &models.Instance{
	//	Name: "instance-33",
	//}
	//
	//ecsSvc := ecs.New(p.session)
	//
	///*
	//	creation
	// */
	//createService, err := createRedisService(instance, ecsSvc, p.provisionerConfig)
	//if err != nil {
	//	p.logger.Error("######## fail createService", zap.Error(err))
	//	return
	//}
	//p.logger.Info("######## success createService", zap.Any("createService", createService))
	//
	///*
	//	status
	// */
	//for {
	//	describeService, err := describePushRedisService(instance, ecsSvc, p.provisionerConfig)
	//	if err != nil {
	//		p.logger.Error("######## fail describeService", zap.Error(err))
	//		return
	//	}
	//	p.logger.Info("######## success describeService", zap.Any("describeService", describeService.Services))
	//	time.Sleep(5 * time.Second)
	//	describeService.Services[0].Count
	//}

}

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

	resultPushRedis, err := p.pushRedisProvisioner.Provision(instance, ecsSvc, serviceDiscoverySvc, p.provisionerConfig)
	if err != nil {
		p.logger.Error("failed while provisioning instance, failed to provision push-redis", zap.Any("instance", instance), zap.Error(err))
		return provisioners.PushServiceProvisionResultFailure
	}

	resultPushStream, err := p.pushStreamProvisioner.Provision(instance, ecsSvc, serviceDiscoverySvc, role, p.provisionerConfig)
	if err != nil {
		p.logger.Error("failed while provisioning instance, failed to provision push-stream", zap.Any("instance", instance))
		return provisioners.PushServiceProvisionResultFailure
	}

	resultPushApi, err := p.pushApiProvisioner.Provision(instance, ecsSvc, serviceDiscoverySvc, ec2Svc, role, p.provisionerConfig)
	if err != nil {
		p.logger.Error("failed while provisioning instance, failed to provision push-api", zap.Any("instance", instance), zap.Error(err))
		return provisioners.PushServiceProvisionResultFailure
	}

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

	session := session.Must(session.NewSession())

	return &ecsProvisioner{
		logger:  logger,
		session: session,
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
