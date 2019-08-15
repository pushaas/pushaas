package ecs_provisioner

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/servicediscovery"

	"github.com/rafaeleyng/pushaas/pushaas/models"
)

const pushApi = "push-api"


type (
	EcsPushApiProvisioner interface {
		Provision(*models.Instance, *ecs.ECS, *servicediscovery.ServiceDiscovery, *ec2.EC2, *iam.GetRoleOutput, *ec2.DescribeNetworkInterfacesOutput, *ecsProvisionerConfig) (*provisionPushApiResult, error)
		DescribeService(*models.Instance, *ecs.ECS, *ecsProvisionerConfig) (*ecs.DescribeServicesOutput, error)
	}

	ecsPushApiProvisioner struct {}
)

func pushApiWithInstance(instanceName string) string {
	return fmt.Sprintf("%s-%s", pushApi, instanceName)
}

/*
	===========================================================================
	provision
	===========================================================================
*/
func (p *ecsPushApiProvisioner) Provision(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	serviceDiscoverySvc *servicediscovery.ServiceDiscovery,
	ec2Svc *ec2.EC2,
	role *iam.GetRoleOutput,
	eni *ec2.DescribeNetworkInterfacesOutput,
	provisionerConfig *ecsProvisionerConfig,
) (*provisionPushApiResult, error) {
	var err error

	taskDefinition, err := p.createPushApiTaskDefinition(instance, ecsSvc, role, eni, provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-api task definition")
	}

	serviceDiscovery, err := p.createPushApiServiceDiscovery(instance, serviceDiscoverySvc, provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-api service discovery service")
	}

	service, err := p.createPushApiService(instance, ecsSvc, serviceDiscovery, provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-api service")
	}

	return &provisionPushApiResult{
		serviceDiscovery: serviceDiscovery,
		taskDefinition:   taskDefinition,
		service:          service,
	}, nil
}

func (p *ecsPushApiProvisioner) createPushApiTaskDefinition(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	role *iam.GetRoleOutput,
	eni *ec2.DescribeNetworkInterfacesOutput,
	provisionerConfig *ecsProvisionerConfig,
) (*ecs.RegisterTaskDefinitionOutput, error) {
	// TODO - technical debt
	pushStreamPublicIp := &eni.NetworkInterfaces[0].Association.PublicIp

	return ecsSvc.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		Family:                  aws.String(pushApiWithInstance(instance.Name)),
		ExecutionRoleArn:        role.Role.Arn,
		NetworkMode:             aws.String(ecs.NetworkModeAwsvpc),
		RequiresCompatibilities: []*string{aws.String(ecs.CompatibilityFargate)},
		Cpu:                     aws.String("256"),
		Memory:                  aws.String("512"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Cpu:               aws.Int64(256),
				Image:             aws.String(provisionerConfig.imagePushApi),
				MemoryReservation: aws.Int64(512),
				Name:              aws.String(pushApi),
				LogConfiguration: &ecs.LogConfiguration{
					LogDriver: aws.String(ecs.LogDriverAwslogs),
					Options: map[string]*string{
						"awslogs-region":        aws.String(provisionerConfig.region),
						"awslogs-group":         aws.String(provisionerConfig.logsGroup),
						"awslogs-stream-prefix": aws.String(provisionerConfig.logsStreamPrefix),
					},
				},
				PortMappings: []*ecs.PortMapping{
					{
						ContainerPort: aws.Int64(8080),
						HostPort:      aws.Int64(8080),
					},
				},
				Environment: []*ecs.KeyValuePair{
					{
						Name:  aws.String("PUSHAPI_REDIS__URL"),
						Value: aws.String(fmt.Sprintf("redis://%s.tsuru:6379", pushRedisWithInstance(instance.Name))),
					},
					{
						Name: aws.String("PUSHAPI_PUSH_STREAM__URL"),
						Value: aws.String(fmt.Sprintf("http://%s.tsuru:9080", pushStreamPublicIp)),
					},
				},
			},
		},
	})
}

func (p *ecsPushApiProvisioner) createPushApiServiceDiscovery(
	instance *models.Instance,
	serviceDiscoverySvc *servicediscovery.ServiceDiscovery,
	provisionerConfig *ecsProvisionerConfig,
) (*servicediscovery.CreateServiceOutput, error) {
	return serviceDiscoverySvc.CreateService(&servicediscovery.CreateServiceInput{
		Name:        aws.String(pushApiWithInstance(instance.Name)),
		NamespaceId: aws.String(provisionerConfig.dnsNamespace),
		DnsConfig: &servicediscovery.DnsConfig{
			DnsRecords: []*servicediscovery.DnsRecord{
				{
					TTL:  aws.Int64(10),
					Type: aws.String("A"),
				},
			},
		},
		HealthCheckCustomConfig: &servicediscovery.HealthCheckCustomConfig{
			FailureThreshold: aws.Int64(1),
		},
	})
}

func (p *ecsPushApiProvisioner) createPushApiService(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	serviceDiscovery *servicediscovery.CreateServiceOutput,
	provisionerConfig *ecsProvisionerConfig,
) (*ecs.CreateServiceOutput, error) {
	return ecsSvc.CreateService(&ecs.CreateServiceInput{
		Cluster:        aws.String(provisionerConfig.cluster),
		DesiredCount:   aws.Int64(1),
		ServiceName:    aws.String(pushApiWithInstance(instance.Name)),
		TaskDefinition: aws.String(pushApiWithInstance(instance.Name)),
		LaunchType:     aws.String(ecs.LaunchTypeFargate),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
				SecurityGroups: []*string{aws.String(provisionerConfig.securityGroup)},
				Subnets:        []*string{aws.String(provisionerConfig.subnet)},
			},
		},
		ServiceRegistries: []*ecs.ServiceRegistry{
			{
				RegistryArn: serviceDiscovery.Service.Arn,
			},
		},
	})
}

/*
	===========================================================================
	deprovision
	===========================================================================
*/
//func deletePushApiService(svc *ecs.ECS) {
//	input := &ecs.DeleteServiceInput{
//		Cluster: aws.String(clusterName),
//		Force: aws.Bool(true),
//		Service: aws.String(pushApiWithInstance),
//	}
//
//	output, err := svc.DeleteService(input)
//	if err != nil {
//		fmt.Println("========== push-api - FAILED DeleteService ==========")
//		panic(err)
//	}
//	fmt.Println("========== push-api - DeleteService ==========")
//	fmt.Println(output.GoString())
//}

/*
	===========================================================================
	other
	===========================================================================
*/
func (p *ecsPushApiProvisioner) DescribeService(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	provisionerConfig *ecsProvisionerConfig,
) (*ecs.DescribeServicesOutput, error) {
	return ecsSvc.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  aws.String(provisionerConfig.cluster),
		Services: []*string{aws.String(pushRedisWithInstance(instance.Name))},
	})
}

func NewEcsPushApiProvisioner() EcsPushApiProvisioner {
	return &ecsPushApiProvisioner{}
}
