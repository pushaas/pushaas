package ecs_provisioner

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/servicediscovery"

	"github.com/rafaeleyng/pushaas/pushaas/models"
)

const pushRedis = "push-redis"


type (
	EcsPushRedisProvisioner interface {
		Provision(*models.Instance, *ecs.ECS, *servicediscovery.ServiceDiscovery, *ecsProvisionerConfig) (*provisionPushRedisResult, error)
		DescribeService(*models.Instance, *ecs.ECS, *ecsProvisionerConfig) (*ecs.DescribeServicesOutput, error)
	}

	ecsPushRedisProvisioner struct {}
)

func pushRedisWithInstance(instanceName string) string {
	return fmt.Sprintf("%s-%s", pushRedis, instanceName)
}

/*
	===========================================================================
	provision
	===========================================================================
*/
func (p *ecsPushRedisProvisioner) Provision(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	serviceDiscoverySvc *servicediscovery.ServiceDiscovery,
	provisionerConfig *ecsProvisionerConfig,
) (*provisionPushRedisResult, error) {
	var err error

	serviceDiscovery, err := p.createRedisServiceDiscovery(instance, serviceDiscoverySvc, provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-redis service discovery service")
	}

	service, err := p.createRedisService(instance, ecsSvc, serviceDiscovery, provisionerConfig)
	if err != nil {
		return nil, errors.New("failed to create push-redis service")
	}

	return &provisionPushRedisResult{
		serviceDiscovery: serviceDiscovery,
		service:          service,
	}, nil
}

func (p *ecsPushRedisProvisioner) createRedisServiceDiscovery(
	instance *models.Instance,
	serviceDiscoverySvc *servicediscovery.ServiceDiscovery,
	provisionerConfig *ecsProvisionerConfig,
) (*servicediscovery.CreateServiceOutput, error) {
	return serviceDiscoverySvc.CreateService(&servicediscovery.CreateServiceInput{
		Name:        aws.String(pushRedisWithInstance(instance.Name)),
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

func (p *ecsPushRedisProvisioner) createRedisService(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	redisDiscovery *servicediscovery.CreateServiceOutput,
	provisionerConfig *ecsProvisionerConfig,
) (*ecs.CreateServiceOutput, error) {
	return ecsSvc.CreateService(&ecs.CreateServiceInput{
		Cluster:        aws.String(provisionerConfig.cluster),
		DesiredCount:   aws.Int64(1),
		ServiceName:    aws.String(pushRedisWithInstance(instance.Name)),
		TaskDefinition: aws.String(pushRedis),
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
				RegistryArn: redisDiscovery.Service.Arn,
			},
		},
	})
}

/*
	===========================================================================
	deprovision
	===========================================================================
*/
//func deleteRedisService(svc *ecs.ECS) {
//	input := &ecs.DeleteServiceInput{
//		Cluster: aws.String(clusterName),
//		Force:   aws.Bool(true),
//		Service: aws.String(pushRedisWithInstance),
//	}
//
//	output, err := svc.DeleteService(input)
//	if err != nil {
//		fmt.Println("========== redis - FAILED DeleteService ==========")
//		panic(err)
//	}
//	fmt.Println("========== redis - DeleteService ==========")
//	fmt.Println(output.GoString())
//}

/*
	===========================================================================
	other
	===========================================================================
*/
func (p *ecsPushRedisProvisioner) DescribeService(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	provisionerConfig *ecsProvisionerConfig,
) (*ecs.DescribeServicesOutput, error) {
	return ecsSvc.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  aws.String(provisionerConfig.cluster),
		Services: []*string{aws.String(pushRedisWithInstance(instance.Name))},
	})
}

func NewEcsPushRedisProvisioner() EcsPushRedisProvisioner {
	return &ecsPushRedisProvisioner{}
}
