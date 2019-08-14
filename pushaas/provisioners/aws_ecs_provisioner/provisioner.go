package aws_ecs_provisioner

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
	awsEcsProvisionerConfig struct {
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

	awsEcsProvisioner struct {
		logger            *zap.Logger
		awsSession        *session.Session
		provisionerConfig *awsEcsProvisionerConfig
	}
)

const ServiceStatusActive = "ACTIVE"
const ServiceStatusDraining = "DRAINING"
const ServiceStatusInactive = "INACTIVE"

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

//func theMain() {
//	const ActionList = "list"
//	const ActionDescribe = "describe"
//	const ActionCreate = "create"
//	const ActionDelete = "delete"
//	action := os.Getenv("ACTION")
//
//	role := getIamRole(iamSvc)
//
//	ignore(ecsSvc, sdSvc, iamSvc, role)
//
//	if action == ActionList {
//		listTaskDescriptions(ecsSvc)
//		listServices(ecsSvc)
//		listPushStreamTasks(ecsSvc)
//		listServiceDiscoveryServices(sdSvc)
//		return
//	}
//
//	if action == ActionDescribe {
//		//describeRedisService(ecsSvc)
//		//describePushApiService(ecsSvc)
//		//describePushStreamService(ecsSvc)
//		//describePushStreamTask(ecsSvc)
//		describePushStreamNetworkInterfaceTask(ecsSvc, ec2Svc)
//		return
//	}
//
//	if action == ActionCreate {
//		redisDiscovery := createRedisServiceDiscovery(sdSvc)
//		createRedisService(ecsSvc, redisDiscovery)
//
//		pushApiDiscovery := createPushApiServiceDiscovery(sdSvc)
//		createPushApiTaskDefinition(ecsSvc, role)
//		createPushApiService(ecsSvc, pushApiDiscovery)
//
//		pushStreamDiscovery := createPushStreamServiceDiscovery(sdSvc)
//		createPushStreamTaskDefinition(ecsSvc, role)
//		createPushStreamService(ecsSvc, pushStreamDiscovery)
//		return
//	}
//
//	if action == ActionDelete {
//		deleteRedisService(ecsSvc)
//		deletePushApiService(ecsSvc)
//		deletePushStreamService(ecsSvc)
//
//		time.Sleep(20 * time.Second)
//
//		deleteServiceDiscoveryServices(sdSvc)
//		return
//	}
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

func (p *awsEcsProvisioner) provisionPushRedis(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	serviceDiscoverySvc *servicediscovery.ServiceDiscovery,
	statusCh chan string,
) (*provisionPushRedisResult, error) {
	var err error

	serviceDiscovery, err := createRedisServiceDiscovery(instance, serviceDiscoverySvc, p.provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-redis service discovery service")
	}

	service, err := createRedisService(instance, ecsSvc, serviceDiscovery, p.provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-redis service")
	}

	monitorRedisServiceStatus(instance, service, ecsSvc, statusCh, p.provisionerConfig)

	return &provisionPushRedisResult{
		serviceDiscovery: serviceDiscovery,
		service:          service,
	}, nil
}

func (p *awsEcsProvisioner) provisionPushStream(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	serviceDiscoverySvc *servicediscovery.ServiceDiscovery,
	role *iam.GetRoleOutput,
) (*provisionPushStreamResult, error) {
	var err error

	taskDefinition, err := createPushStreamTaskDefinition(instance, ecsSvc, role, p.provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-stream task definition")
	}

	serviceDiscovery, err := createPushStreamServiceDiscovery(instance, serviceDiscoverySvc, p.provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-stream service discovery service")
	}

	service, err := createPushStreamService(instance, ecsSvc, serviceDiscovery, p.provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-stream service")
	}

	return &provisionPushStreamResult{
		serviceDiscovery: serviceDiscovery,
		taskDefinition:   taskDefinition,
		service:          service,
	}, nil
}

func (p *awsEcsProvisioner) provisionPushApi(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	serviceDiscoverySvc *servicediscovery.ServiceDiscovery,
	ec2Svc *ec2.EC2,
	role *iam.GetRoleOutput,
) (*provisionPushApiResult, error) {
	var err error

	eni, err := describePushStreamNetworkInterfaceTask(instance, ecsSvc, ec2Svc, p.provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to obtain push-stream public IP to create push-api task definition")
	}

	taskDefinition, err := createPushApiTaskDefinition(instance, ecsSvc, role, eni, p.provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-api task definition")
	}

	serviceDiscovery, err := createPushApiServiceDiscovery(instance, serviceDiscoverySvc, p.provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-api service discovery service")
	}

	service, err := createPushApiService(instance, ecsSvc, serviceDiscovery, p.provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-api service")
	}

	return &provisionPushApiResult{
		serviceDiscovery: serviceDiscovery,
		taskDefinition:   taskDefinition,
		service:          service,
	}, nil
}

func (p *awsEcsProvisioner) Provision(instance *models.Instance) provisioners.ProvisionResult {
	p.logger.Info("starting provision for instance", zap.Any("instance", instance))

	iamSvc := iam.New(p.awsSession)
	ecsSvc := ecs.New(p.awsSession)
	ec2Svc := ec2.New(p.awsSession)
	serviceDiscoverySvc := servicediscovery.New(p.awsSession)

	var err error
	role, err := getIamRole(iamSvc)
	if err != nil {
		p.logger.Error("failed while provisioning instance, failed to get iam role", zap.Any("instance", instance), zap.Error(err))
		return provisioners.ProvisionResultFailure
	}

	chRedis := make(chan string)
	resultPushRedis, err := p.provisionPushRedis(instance, ecsSvc, serviceDiscoverySvc, chRedis)
	if err != nil {
		p.logger.Error("failed while provisioning instance, failed to provision push-redis", zap.Any("instance", instance), zap.Error(err))
		return provisioners.ProvisionResultFailure
	}
	redisStatus := <- chRedis
	p.logger.Info("################## redis", zap.String("redisStatus", redisStatus))
	//if redisStatus != ServiceStatusActive {
	//
	//}

	resultPushStream, err := p.provisionPushStream(instance, ecsSvc, serviceDiscoverySvc, role)
	if err != nil {
		p.logger.Error("failed while provisioning instance, failed to provision push-stream", zap.Any("instance", instance))
		return provisioners.ProvisionResultFailure
	}

	resultPushApi, err := p.provisionPushApi(instance, ecsSvc, serviceDiscoverySvc, ec2Svc, role)
	if err != nil {
		p.logger.Error("failed while provisioning instance, failed to provision push-api", zap.Any("instance", instance), zap.Error(err))
		return provisioners.ProvisionResultFailure
	}

	p.logger.Info(
		"finishing provision for instance",
		zap.Any("instance", instance),
		zap.Any("resultPushRedis", resultPushRedis),
		zap.Any("resultPushStream", resultPushStream),
		zap.Any("resultPushApi", resultPushApi),
	)

	return provisioners.ProvisionResultSuccess
}

func (p *awsEcsProvisioner) deprovisionRedis(instance *models.Instance) {

}

func (p *awsEcsProvisioner) Deprovision(instance *models.Instance) provisioners.DeprovisionResult {
	p.logger.Info("starting deprovision for instance", zap.Any("instance", instance))

	return provisioners.DeprovisionResultSuccess
}

func NewAwsEcsProvisioner(config *viper.Viper, logger *zap.Logger) (provisioners.Provisioner, error) {
	imagePushApi := config.GetString("provisioner.aws.image-push-api")
	imagePushAgent := config.GetString("provisioner.aws.image-push-agent")
	imagePushStream := config.GetString("provisioner.aws.image-push-stream")

	region := config.GetString("provisioner.aws.region")
	cluster := config.GetString("provisioner.aws.cluster")
	logsStreamPrefix := config.GetString("provisioner.aws.logs-stream-prefix")
	logsGroup := config.GetString("provisioner.aws.logs-group")
	// TODO comes from `scripts/40-pushaas/30-create-cluster/terraform.tfstate`, should create specific for each part of push service
	securityGroup := config.GetString("provisioner.aws.security-group")
	// TODO comes from `scripts/10-vpc/10-create-vpc/terraform.tfstate` - should have public and private
	subnet := config.GetString("provisioner.aws.subnet")
	// TODO comes from `scripts/30-dns/10-create-namespace/terraform.tfstate`
	dnsNamespace := config.GetString("provisioner.aws.dns-namespace")

	requiredVars := []string{
		securityGroup,
		subnet,
		dnsNamespace,
	}

	for _, v := range requiredVars {
		if v == "" {
			return nil, errors.New(fmt.Sprintf("awsEcsProvisioner env required and not set: %s", v))
		}
	}

	awsSession := session.Must(session.NewSession())

	return &awsEcsProvisioner{
		logger:     logger,
		awsSession: awsSession,
		provisionerConfig: &awsEcsProvisionerConfig{
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
	}, nil
}
