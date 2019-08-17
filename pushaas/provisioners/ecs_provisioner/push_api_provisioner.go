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

const pushApi = "push-api"


type (
	EcsPushApiProvisioner interface {
		Provision(*models.Instance, *iam.GetRoleOutput, *ec2.DescribeNetworkInterfacesOutput, chan *provisionPushApiResult)
	}

	ecsPushApiProvisioner struct {
		logger            *zap.Logger
		provisionerConfig *EcsProvisionerConfig
	}

	provisionPushApiResult struct {
		service          *ecs.CreateServiceOutput
		serviceDiscovery *servicediscovery.CreateServiceOutput
		taskDefinition   *ecs.RegisterTaskDefinitionOutput
		err              error
	}
)

func pushApiWithInstance(instanceName string) string {
	return fmt.Sprintf("%s-%s", pushApi, instanceName)
}

/*
	===========================================================================
	provision
	===========================================================================
*/
func (p *ecsPushApiProvisioner) Provision(instance *models.Instance, role *iam.GetRoleOutput, eni *ec2.DescribeNetworkInterfacesOutput, ch chan *provisionPushApiResult) {
	var err error

	// create task definition
	taskDefinition, err := p.createPushApiTaskDefinition(instance, role, eni)
	if err != nil {
		ch <- &provisionPushApiResult{err: errors.New("failed to create push-api task definition")}
		return
	}
	p.logger.Debug("[push-api] did create task definition")

	// create service discovery
	serviceDiscovery, err := p.createPushApiServiceDiscovery(instance)
	if err != nil {
		ch <- &provisionPushApiResult{err: errors.New("failed to create push-api service discovery service")}
		return
	}
	p.logger.Debug("[push-api] did create service discovery")

	// create service
	service, err := p.createPushApiService(instance, serviceDiscovery)
	if err != nil {
		ch <- &provisionPushApiResult{err: errors.New("failed to create push-api service")}
		return
	}
	p.logger.Debug("[push-api] did create service")

	// wait for service
	waitCh := make(chan bool)
	go waitServiceUp(instance, waitCh, p.describeService)
	if serviceUp := <-waitCh; !serviceUp {
		ch <- &provisionPushApiResult{err: errors.New("push-api service did not become available")}
		return
	}
	p.logger.Debug("[push-api] service is up")

	ch <- &provisionPushApiResult{
		service:          service,
		serviceDiscovery: serviceDiscovery,
		taskDefinition:   taskDefinition,
	}
}

func (p *ecsPushApiProvisioner) createPushApiTaskDefinition(
	instance *models.Instance,
	role *iam.GetRoleOutput,
	eni *ec2.DescribeNetworkInterfacesOutput,
) (*ecs.RegisterTaskDefinitionOutput, error) {
	// TODO technical debt
	pushStreamPublicIp := *eni.NetworkInterfaces[0].Association.PublicIp

	return p.provisionerConfig.ecs.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		Family:                  aws.String(pushApiWithInstance(instance.Name)),
		ExecutionRoleArn:        role.Role.Arn,
		NetworkMode:             aws.String(ecs.NetworkModeAwsvpc),
		RequiresCompatibilities: []*string{aws.String(ecs.CompatibilityFargate)},
		Cpu:                     aws.String("256"),
		Memory:                  aws.String("512"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Cpu:               aws.Int64(256),
				Image:             p.provisionerConfig.imagePushApi,
				MemoryReservation: aws.Int64(512),
				Name:              aws.String(pushApi),
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
) (*servicediscovery.CreateServiceOutput, error) {
	return p.provisionerConfig.serviceDiscovery.CreateService(&servicediscovery.CreateServiceInput{
		Name:        aws.String(pushApiWithInstance(instance.Name)),
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

func (p *ecsPushApiProvisioner) createPushApiService(
	instance *models.Instance,
	serviceDiscovery *servicediscovery.CreateServiceOutput,
) (*ecs.CreateServiceOutput, error) {
	return p.provisionerConfig.ecs.CreateService(&ecs.CreateServiceInput{
		Cluster:        p.provisionerConfig.cluster,
		DesiredCount:   aws.Int64(1),
		ServiceName:    aws.String(pushApiWithInstance(instance.Name)),
		TaskDefinition: aws.String(pushApiWithInstance(instance.Name)),
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
func (p *ecsPushApiProvisioner) describeService(
	instance *models.Instance,
) (*ecs.DescribeServicesOutput, error) {
	return p.provisionerConfig.ecs.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  p.provisionerConfig.cluster,
		Services: []*string{aws.String(pushRedisWithInstance(instance.Name))},
	})
}

func NewEcsPushApiProvisioner(logger *zap.Logger, provisionerConfig *EcsProvisionerConfig) EcsPushApiProvisioner {
	return &ecsPushApiProvisioner{
		logger:            logger,
		provisionerConfig: provisionerConfig,
	}
}
