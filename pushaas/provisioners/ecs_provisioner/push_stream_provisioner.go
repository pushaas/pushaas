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

const pushAgent = "push-agent"
const pushStream = "push-stream"

type (
	EcsPushStreamProvisioner interface {
		Provision(*models.Instance, *ecs.ECS, *servicediscovery.ServiceDiscovery, *iam.GetRoleOutput, *ecsProvisionerConfig) (*provisionPushStreamResult, error)
	}

	ecsPushStreamProvisioner struct {}
)

func pushStreamWithInstance(instanceName string) string {
	return fmt.Sprintf("%s-%s", pushStream, instanceName)
}

/*
	===========================================================================
	provision
	===========================================================================
*/
func (e *ecsPushStreamProvisioner) Provision(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	serviceDiscoverySvc *servicediscovery.ServiceDiscovery,
	role *iam.GetRoleOutput,
	provisionerConfig *ecsProvisionerConfig,
) (*provisionPushStreamResult, error) {
	var err error

	taskDefinition, err := createPushStreamTaskDefinition(instance, ecsSvc, role, provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-stream task definition")
	}

	serviceDiscovery, err := createPushStreamServiceDiscovery(instance, serviceDiscoverySvc, provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-stream service discovery service")
	}

	service, err := createPushStreamService(instance, ecsSvc, serviceDiscovery, provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-stream service")
	}

	return &provisionPushStreamResult{
		serviceDiscovery: serviceDiscovery,
		taskDefinition:   taskDefinition,
		service:          service,
	}, nil
}

func createPushStreamTaskDefinition(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	role *iam.GetRoleOutput,
	provisionerConfig *ecsProvisionerConfig,
) (*ecs.RegisterTaskDefinitionOutput, error) {
	return ecsSvc.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		Family:                  aws.String(pushStreamWithInstance(instance.Name)),
		ExecutionRoleArn:        role.Role.Arn,
		NetworkMode:             aws.String(ecs.NetworkModeAwsvpc),
		RequiresCompatibilities: []*string{aws.String(ecs.CompatibilityFargate)},
		Cpu:                     aws.String("512"),
		Memory:                  aws.String("1024"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Name:              aws.String(pushStream),
				Cpu:               aws.Int64(256),
				Image:             aws.String(provisionerConfig.imagePushStream),
				MemoryReservation: aws.Int64(512),
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
						ContainerPort: aws.Int64(9080),
						HostPort:      aws.Int64(9080),
					},
				},
			},
			{
				Name:  aws.String(pushAgent),
				Cpu:   aws.Int64(256),
				Image: aws.String(provisionerConfig.imagePushAgent),
				DependsOn: []*ecs.ContainerDependency{
					{
						Condition:     aws.String(ecs.ContainerConditionStart),
						ContainerName: aws.String(pushStream),
					},
				},
				MemoryReservation: aws.Int64(512),
				LogConfiguration: &ecs.LogConfiguration{
					LogDriver: aws.String(ecs.LogDriverAwslogs),
					Options: map[string]*string{
						"awslogs-region":        aws.String(provisionerConfig.region),
						"awslogs-group":         aws.String(provisionerConfig.logsGroup),
						"awslogs-stream-prefix": aws.String(provisionerConfig.logsStreamPrefix),
					},
				},
				Environment: []*ecs.KeyValuePair{
					{
						Name:  aws.String("PUSHAGENT_REDIS__URL"),
						Value: aws.String("redis://" + pushRedisWithInstance(instance.Name) + ".tsuru:6379"),
					},
					{
						Name:  aws.String("PUSHAGENT_PUSH_STREAM__URL"),
						Value: aws.String("http://" + pushStreamWithInstance(instance.Name) + ".tsuru:9080"),
					},
				},
			},
		},
	})
}

func createPushStreamServiceDiscovery(
	instance *models.Instance,
	serviceDiscoverySvc *servicediscovery.ServiceDiscovery,
	provisionerConfig *ecsProvisionerConfig,
) (*servicediscovery.CreateServiceOutput, error) {
	return serviceDiscoverySvc.CreateService(&servicediscovery.CreateServiceInput{
		Name:        aws.String(pushStreamWithInstance(instance.Name)),
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

func createPushStreamService(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	pushStreamDiscovery *servicediscovery.CreateServiceOutput,
	provisionerConfig *ecsProvisionerConfig,
) (*ecs.CreateServiceOutput, error) {
	return ecsSvc.CreateService(&ecs.CreateServiceInput{
		Cluster:        aws.String(provisionerConfig.cluster),
		DesiredCount:   aws.Int64(1),
		ServiceName:    aws.String(pushStreamWithInstance(instance.Name)),
		TaskDefinition: aws.String(pushStreamWithInstance(instance.Name)),
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
				RegistryArn: pushStreamDiscovery.Service.Arn,
			},
		},
	})
}

/*
	===========================================================================
	deprovision
	===========================================================================
*/
//func deletePushStreamService(svc *ecs.ECS) {
//	input := &ecs.DeleteServiceInput{
//		Cluster: aws.String(clusterName),
//		Force:   aws.Bool(true),
//		Service: aws.String(pushStreamWithInstance(instance.Name)),
//	}
//
//	output, err := svc.DeleteService(input)
//	if err != nil {
//		fmt.Println("========== push-stream - FAILED DeleteService ==========")
//		panic(err)
//	}
//	fmt.Println("========== push-stream - DeleteService ==========")
//	fmt.Println(output.GoString())
//}

/*
	===========================================================================
	other
	===========================================================================
*/
func listPushStreamTasks(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	provisionerConfig *ecsProvisionerConfig,
) (*ecs.ListTasksOutput, error) {
	return ecsSvc.ListTasks(&ecs.ListTasksInput{
		Cluster:     aws.String(provisionerConfig.cluster),
		ServiceName: aws.String(pushStreamWithInstance(instance.Name)),
	})
}

func describePushStreamTasks(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	provisionerConfig *ecsProvisionerConfig,
) (*ecs.DescribeTasksOutput, error) {
	listOutput, err := listPushStreamTasks(instance, ecsSvc, provisionerConfig)
	if err != nil {
		return nil, err
	}

	return ecsSvc.DescribeTasks(&ecs.DescribeTasksInput{
		Tasks:   []*string{listOutput.TaskArns[0]},
		Cluster: aws.String(provisionerConfig.cluster),
	})
}

func describePushStreamNetworkInterfaceTask(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	ec2Svc *ec2.EC2,
	provisionerConfig *ecsProvisionerConfig,
) (*ec2.DescribeNetworkInterfacesOutput, error) {
	describeOutput, _ := describePushStreamTasks(instance, ecsSvc, provisionerConfig)

	var eniId *string
	for _, kv := range describeOutput.Tasks[0].Attachments[0].Details {
		if *kv.Name == "networkInterfaceId" {
			eniId = kv.Value
		}
	}

	return ec2Svc.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{eniId},
	})
}

//func describePushStreamService(svc *ecs.ECS) {
//	input := &ecs.DescribeServicesInput{
//		Cluster:  aws.String(clusterName),
//		Services: []*string{aws.String(pushStreamWithInstance(instance.Name))},
//	}
//
//	output, err := svc.DescribeServices(input)
//	if err != nil {
//		fmt.Println("========== push-stream - FAILED DescribeServices ==========")
//		panic(err)
//	}
//	fmt.Println("========== push-stream - DescribeServices ==========")
//	fmt.Println(output.GoString())
//}

func NewEcsPushStreamProvisioner() EcsPushStreamProvisioner {
	return &ecsPushStreamProvisioner{}
}
