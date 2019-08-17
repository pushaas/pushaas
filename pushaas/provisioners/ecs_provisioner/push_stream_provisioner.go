package ecs_provisioner

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/models"
)

const pushAgent = "push-agent"
const pushStream = "push-stream"

type (
	EcsPushStreamProvisioner interface {
		Provision(*models.Instance, *iam.GetRoleOutput, chan *provisionPushStreamResult)
	}

	ecsPushStreamProvisioner struct{
		logger            *zap.Logger
		provisionerConfig *EcsProvisionerConfig
	}

	provisionPushStreamResult struct {
		eni              *ec2.DescribeNetworkInterfacesOutput
		service          *ecs.CreateServiceOutput
		serviceDiscovery *servicediscovery.CreateServiceOutput
		taskDefinition   *ecs.RegisterTaskDefinitionOutput
		err              error
	}
)

func pushStreamWithInstance(instanceName string) string {
	return fmt.Sprintf("%s-%s", pushStream, instanceName)
}

/*
	===========================================================================
	provision
	===========================================================================
*/
func (p *ecsPushStreamProvisioner) Provision(instance *models.Instance, role *iam.GetRoleOutput, ch chan *provisionPushStreamResult) {
	var err error

	// create task definition
	taskDefinition, err := p.createPushStreamTaskDefinition(instance, role)
	if err != nil {
		ch <- &provisionPushStreamResult{err: errors.New("failed to create push-stream task definition")}
		return
	}
	p.logger.Debug("[push-stream] did create task definition")

	// create service discovery
	serviceDiscovery, err := p.createPushStreamServiceDiscovery(instance)
	if err != nil {
		ch <- &provisionPushStreamResult{err: errors.New("failed to create push-stream service discovery service")}
		return
	}
	p.logger.Debug("[push-stream] did create service discovery")

	// create service
	service, err := p.createPushStreamService(instance, serviceDiscovery)
	if err != nil {
		ch <- &provisionPushStreamResult{err: errors.New("failed to create push-stream service")}
		return
	}
	p.logger.Debug("[push-stream] did create service")

	// wait for service
	waitCh := make(chan bool)
	go waitServiceUp(instance, waitCh, p.describeService)
	if serviceUp := <-waitCh; !serviceUp {
		ch <- &provisionPushStreamResult{err: errors.New("push-stream service did not become available")}
		return
	}
	p.logger.Debug("[push-stream] service is up")

	// wait for network interface
	eniCh := make(chan bool)
	go waitTaskNetworkInterface(instance, eniCh, p.describePushStreamTaskNetworkInterface)
	if isEniUp := <-eniCh; !isEniUp {
		ch <- &provisionPushStreamResult{err: errors.New("push-stream ENI failed to become available")}
		return
	}

	// get network interface
	eni, err := p.describePushStreamTaskNetworkInterface(instance)
	if err != nil {
		ch <- &provisionPushStreamResult{err: errors.New("push-stream ENI could not be retrieved")}
		return
	}
	p.logger.Debug("[push-stream] network interface is up")

	ch <- &provisionPushStreamResult{
		eni:              eni,
		service:          service,
		serviceDiscovery: serviceDiscovery,
		taskDefinition:   taskDefinition,
	}
}

func (p *ecsPushStreamProvisioner) createPushStreamTaskDefinition(instance *models.Instance, role *iam.GetRoleOutput) (*ecs.RegisterTaskDefinitionOutput, error) {
	return p.provisionerConfig.ecs.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
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
				Image:             p.provisionerConfig.imagePushStream,
				MemoryReservation: aws.Int64(512),
				LogConfiguration: &ecs.LogConfiguration{
					LogDriver: aws.String(ecs.LogDriverAwslogs),
					Options: map[string]*string{
						"awslogs-region":        p.provisionerConfig.region,
						"awslogs-group":         p.provisionerConfig.logsGroup,
						"awslogs-stream-prefix": p.provisionerConfig.logsStreamPrefix,
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
				Image: p.provisionerConfig.imagePushAgent,
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
						"awslogs-region":        p.provisionerConfig.region,
						"awslogs-group":         p.provisionerConfig.logsGroup,
						"awslogs-stream-prefix": p.provisionerConfig.logsStreamPrefix,
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

func (p *ecsPushStreamProvisioner) createPushStreamServiceDiscovery(instance *models.Instance) (*servicediscovery.CreateServiceOutput, error) {
	return p.provisionerConfig.serviceDiscovery.CreateService(&servicediscovery.CreateServiceInput{
		Name:        aws.String(pushStreamWithInstance(instance.Name)),
		NamespaceId: p.provisionerConfig.dnsNamespace,
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

func (p *ecsPushStreamProvisioner) createPushStreamService(instance *models.Instance, pushStreamDiscovery *servicediscovery.CreateServiceOutput) (*ecs.CreateServiceOutput, error) {
	return p.provisionerConfig.ecs.CreateService(&ecs.CreateServiceInput{
		Cluster:        p.provisionerConfig.cluster,
		DesiredCount:   aws.Int64(1),
		ServiceName:    aws.String(pushStreamWithInstance(instance.Name)),
		TaskDefinition: aws.String(pushStreamWithInstance(instance.Name)),
		LaunchType:     aws.String(ecs.LaunchTypeFargate),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
				SecurityGroups: []*string{p.provisionerConfig.securityGroup},
				Subnets:        []*string{p.provisionerConfig.subnet},
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
func (p *ecsPushStreamProvisioner) listPushStreamTasks(instance *models.Instance) (*ecs.ListTasksOutput, error) {
	return p.provisionerConfig.ecs.ListTasks(&ecs.ListTasksInput{
		Cluster:     p.provisionerConfig.cluster,
		ServiceName: aws.String(pushStreamWithInstance(instance.Name)),
	})
}

func (p *ecsPushStreamProvisioner) describePushStreamTasks(instance *models.Instance) (*ecs.DescribeTasksOutput, error) {
	listOutput, err := p.listPushStreamTasks(instance)
	if err != nil {
		return nil, err
	}

	if len(listOutput.TaskArns) == 0 {
		return nil, errors.New(fmt.Sprintf("[describePushStreamTasks] no tasks in service %s", pushStreamWithInstance(instance.Name)))
	}

	return p.provisionerConfig.ecs.DescribeTasks(&ecs.DescribeTasksInput{
		Tasks:   []*string{listOutput.TaskArns[0]},
		Cluster: p.provisionerConfig.cluster,
	})
}

func (p *ecsPushStreamProvisioner) describePushStreamTaskNetworkInterface(instance *models.Instance) (*ec2.DescribeNetworkInterfacesOutput, error) {
	describeOutput, err := p.describePushStreamTasks(instance)
	if err != nil {
		return nil, err
	}

	if len(describeOutput.Tasks) == 0 || len(describeOutput.Tasks[0].Attachments) == 0 {
		return nil, errors.New(fmt.Sprintf("[describePushStreamTaskNetworkInterface] no tasks or attachments found for service %s", pushStreamWithInstance(instance.Name)))
	}

	var eniId *string
	for _, kv := range describeOutput.Tasks[0].Attachments[0].Details {
		if *kv.Name == "networkInterfaceId" {
			eniId = kv.Value
		}
	}

	return p.provisionerConfig.ec2.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{eniId},
	})
}

func (p *ecsPushStreamProvisioner) describeService(instance *models.Instance) (*ecs.DescribeServicesOutput, error) {
	return p.provisionerConfig.ecs.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  p.provisionerConfig.cluster,
		Services: []*string{aws.String(pushStreamWithInstance(instance.Name))},
	})
}

func NewEcsPushStreamProvisioner(logger *zap.Logger, provisionerConfig *EcsProvisionerConfig) EcsPushStreamProvisioner {
	return &ecsPushStreamProvisioner{
		logger:            logger,
		provisionerConfig: provisionerConfig,
	}
}
