package aws_ecs_provisioner

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
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

//func ignore(
//	ecsSvc *ecs.ECS,
//	discovery *servicediscovery.ServiceDiscovery,
//	iamSvc *iam.IAM,
//	output *iam.GetRoleOutput,
//) {}
//
//func theMain() {
//	const ActionList = "list"
//	const ActionDescribe = "describe"
//	const ActionCreate = "create"
//	const ActionDelete = "delete"
//	action := os.Getenv("ACTION")
//
//	roleOutput := getIamRole(iamSvc)
//
//	ignore(ecsSvc, sdSvc, iamSvc, roleOutput)
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
//		createPushApiTaskDefinition(ecsSvc, roleOutput)
//		createPushApiService(ecsSvc, pushApiDiscovery)
//
//		pushStreamDiscovery := createPushStreamServiceDiscovery(sdSvc)
//		createPushStreamTaskDefinition(ecsSvc, roleOutput)
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

/*
	===========================================================================
	implementation
	===========================================================================
*/

func (p *awsEcsProvisioner) provisionRedis(instance *models.Instance, ecsSvc *ecs.ECS, serviceDiscoverySvc *servicediscovery.ServiceDiscovery) error {
	var err error

	serviceDiscovery, err := createRedisServiceDiscovery(instance, serviceDiscoverySvc, p.provisionerConfig)
	if err != nil {
		return errors.New("failed to create redis service discovery service")
	}

	service, err := createRedisService(instance, ecsSvc, serviceDiscovery, p.provisionerConfig)
	if err != nil {
		return errors.New("failed to create redis service")
	}

	return nil
}

func (p *awsEcsProvisioner) provisionPushApi(instance *models.Instance, ecsSvc *ecs.ECS, serviceDiscoverySvc *servicediscovery.ServiceDiscovery, role *iam.GetRoleOutput) error {
	var err error

	serviceDiscovery, err := createPushApiServiceDiscovery(instance, serviceDiscoverySvc, p.provisionerConfig)
	if err != nil {
		return errors.New("failed to create api service discovery service")
	}

	taskDefinition, err := createPushApiTaskDefinition(instance, ecsSvc, role, p.provisionerConfig)
	if err != nil {
		return errors.New("failed to create api task definition")
	}

	service, err := createPushApiService(instance, ecsSvc, serviceDiscovery, p.provisionerConfig)
	if err != nil {
		return errors.New("failed to create api service")
	}

	return nil
}

func (p *awsEcsProvisioner) Provision(instance *models.Instance) provisioners.ProvisionResult {
	var err error
	p.logger.Info("######## did call Provision")

	iamSvc := iam.New(p.awsSession)
	ecsSvc := ecs.New(p.awsSession)
	serviceDiscoverySvc := servicediscovery.New(p.awsSession)
	//ec2Svc := ec2.New(p.awsSession)

	role, err := getIamRole(iamSvc)
	if err != nil {
		return provisioners.ProvisionResultFailure
	}

	err = p.provisionRedis(instance, ecsSvc, serviceDiscoverySvc)
	if err != nil {
		return provisioners.ProvisionResultFailure
	}

	err = p.provisionPushApi(instance, ecsSvc, serviceDiscoverySvc, role)
	if err != nil {
		return provisioners.ProvisionResultFailure
	}

	return provisioners.ProvisionResultSuccess
}

func (p *awsEcsProvisioner) deprovisionRedis(instance *models.Instance) {

}

func (p *awsEcsProvisioner) Deprovision(instance *models.Instance) provisioners.DeprovisionResult {
	p.logger.Info("######## did call Deprovision")

	p.deprovisionRedis(instance)

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
