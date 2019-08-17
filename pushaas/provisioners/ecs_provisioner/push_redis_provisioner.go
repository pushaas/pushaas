package ecs_provisioner

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/models"
)

const pushRedis = "push-redis"

type (
	EcsPushRedisProvisioner interface {
		Provision(*models.Instance, chan *provisionPushRedisResult)
		//Deprovision(*models.Instance, func(*models.Instance, chan bool, func(*models.Instance) (*ecs.DescribeServicesOutput, error))) (*deprovisionPushRedisResult, error)
	}

	ecsPushRedisProvisioner struct {
		logger            *zap.Logger
		provisionerConfig *EcsProvisionerConfig
	}

	provisionPushRedisResult struct {
		service          *ecs.CreateServiceOutput
		serviceDiscovery *servicediscovery.CreateServiceOutput
		err              error
	}

	//deprovisionPushRedisResult struct {
	//	service          *ecs.DeleteServiceOutput
	//	serviceDiscovery *servicediscovery.DeleteServiceOutput
	//}
)

func pushRedisWithInstance(instanceName string) string {
	return fmt.Sprintf("%s-%s", pushRedis, instanceName)
}

/*
	===========================================================================
	provision
	===========================================================================
*/
func (p *ecsPushRedisProvisioner) Provision(instance *models.Instance, ch chan *provisionPushRedisResult) {
	var err error

	// create service discovery
	serviceDiscovery, err := p.createRedisServiceDiscovery(instance)
	if err != nil {
		ch <- &provisionPushRedisResult{err: errors.New("failed to create push-redis service discovery service")}
		return
	}
	p.logger.Debug("[push-redis] did create service discovery")

	// create service
	service, err := p.createRedisService(instance, serviceDiscovery)
	if err != nil {
		ch <- &provisionPushRedisResult{err: errors.New("failed to create push-redis service")}
		return
	}
	p.logger.Debug("[push-redis] did create service")

	// wait for service
	waitCh := make(chan bool)
	go waitServiceUp(instance, waitCh, p.describeService)
	if serviceUp := <-waitCh; !serviceUp {
		ch <- &provisionPushRedisResult{err: errors.New("push-redis service did not become available")}
		return
	}
	p.logger.Debug("[push-redis] service is up")

	ch <- &provisionPushRedisResult{
		service:          service,
		serviceDiscovery: serviceDiscovery,
	}
}

func (p *ecsPushRedisProvisioner) createRedisServiceDiscovery(instance *models.Instance) (*servicediscovery.CreateServiceOutput, error) {
	return p.provisionerConfig.serviceDiscovery.CreateService(&servicediscovery.CreateServiceInput{
		Name:        aws.String(pushRedisWithInstance(instance.Name)),
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

func (p *ecsPushRedisProvisioner) createRedisService(instance *models.Instance,  serviceDiscovery *servicediscovery.CreateServiceOutput) (*ecs.CreateServiceOutput, error) {
	return p.provisionerConfig.ecs.CreateService(&ecs.CreateServiceInput{
		Cluster:        p.provisionerConfig.cluster,
		DesiredCount:   aws.Int64(1),
		ServiceName:    aws.String(pushRedisWithInstance(instance.Name)),
		TaskDefinition: aws.String(pushRedis),
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
//func (p *ecsPushRedisProvisioner) Deprovision(
//	instance *models.Instance,
//	ecsSvc *ecs.ECS,
//	serviceDiscoverySvc *servicediscovery.ServiceDiscovery,
//	waitServiceDown func(*models.Instance, *ecs.ECS, chan bool, func(*models.Instance, *ecs.ECS) (*ecs.DescribeServicesOutput, error),
//	)) (*deprovisionPushRedisResult, error) {
//	//var err error
//	//
//	//service, err := p.deleteRedisService(instance, ecsSvc, provisionerConfig)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	//chRedis
//	//
//	//serviceDiscovery, err := p.deleteRedisServiceDiscovery(instance, serviceDiscoverySvc, provisionerConfig)
//	//fmt.Println("############### rafael START")
//	//fmt.Println(serviceDiscovery)
//	//fmt.Println(err)
//	//fmt.Println("############### rafael END")
//	//if err != nil {
//	//	return nil, err
//	//}
//
//	return &deprovisionPushRedisResult{
//		//serviceDiscovery: serviceDiscovery,
//		//service:          service,
//	}, nil
//}

/*
func (p *ecsPushRedisProvisioner) deleteRedisService(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	provisionerConfig *ecsProvisionerConfig,
) (*ecs.DeleteServiceOutput, error) {
	return ecsSvc.DeleteService(&ecs.DeleteServiceInput{
		Cluster: aws.String(p.provisionerConfig.cluster),
		Force:   aws.Bool(true),
		Service: aws.String(pushRedisWithInstance(instance.Name)),
	})
}

func (p *ecsPushRedisProvisioner) deleteRedisServiceDiscovery(
	instance *models.Instance,
	serviceDiscoverySvc *servicediscovery.ServiceDiscovery,
	provisionerConfig *ecsProvisionerConfig,
) (*servicediscovery.DeleteServiceOutput, error) {
	listServiceResult, err := p.listServiceDiscoveryServices(serviceDiscoverySvc)
	if err != nil {
		return nil, err
	}

	for _, service := range listServiceResult.Services {
		if *service.Name == pushRedisWithInstance(instance.Name) {
			return serviceDiscoverySvc.DeleteService(&servicediscovery.DeleteServiceInput{
				Id: service.Id,
			})
		}
	}
	return nil, errors.New(fmt.Sprintf("could not find push-redis service discovery service for instance %s", instance.Name))
}
*/

/*
	===========================================================================
	other
	===========================================================================
*/
func (p *ecsPushRedisProvisioner) listServiceDiscoveryServices() (*servicediscovery.ListServicesOutput, error) {
	return p.provisionerConfig.serviceDiscovery.ListServices(&servicediscovery.ListServicesInput{})
}

func (p *ecsPushRedisProvisioner) describeService(instance *models.Instance) (*ecs.DescribeServicesOutput, error) {
	return p.provisionerConfig.ecs.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  p.provisionerConfig.cluster,
		Services: []*string{aws.String(pushRedisWithInstance(instance.Name))},
	})
}

func NewEcsPushRedisProvisioner(logger *zap.Logger, provisionerConfig *EcsProvisionerConfig) EcsPushRedisProvisioner {
	return &ecsPushRedisProvisioner{
		logger:            logger,
		provisionerConfig: provisionerConfig,
	}
}
