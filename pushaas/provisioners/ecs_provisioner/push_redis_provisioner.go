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
		Deprovision(*models.Instance, chan *deprovisionPushRedisResult)
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

	deprovisionPushRedisResult struct {
		service          *ecs.DeleteServiceOutput
		serviceDiscovery *servicediscovery.DeleteServiceOutput
		err              error
	}
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
		//p.logger.Error()
		ch <- &provisionPushRedisResult{err: err}
		return
	}
	p.logger.Debug("[push-redis] did create service discovery")

	// create service
	service, err := p.createRedisService(instance, serviceDiscovery)
	if err != nil {
		ch <- &provisionPushRedisResult{err: err}
		return
	}
	p.logger.Debug("[push-redis] did create service")

	// wait for service to go up
	waitCh := make(chan bool)
	go waitServiceUp(p.logger, instance, waitCh, p.describeService)
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
func (p *ecsPushRedisProvisioner) Deprovision(instance *models.Instance, ch chan *deprovisionPushRedisResult) {
	var err error

	// get service
	describedService, err := p.describeService(instance)
	if err != nil {
		ch <- &deprovisionPushRedisResult{err: err}
		return
	}
	if len(describedService.Services) == 0 {
		ch <- &deprovisionPushRedisResult{err: errors.New(fmt.Sprintf("[push-redis] could  not find service %s", pushRedisWithInstance(instance.Name)))}
		return
	}
	p.logger.Debug("[push-redis] did locate service")

	// scale to 0 tasks
	_, err = p.stopService(instance, describedService)
	if err != nil {
		ch <- &deprovisionPushRedisResult{err: err}
		return
	}
	p.logger.Debug("[push-redis] did update service to desiredCount 0")

	// wait tasks to stop
	waitTasksCh := make(chan bool)
	go waitServiceStopAllTasks(p.logger, instance, waitTasksCh, p.describeService)
	if serviceDown := <-waitTasksCh; !serviceDown {
		ch <- &deprovisionPushRedisResult{err: errors.New("[push-redis] service did not remove all tasks")}
		return
	}
	p.logger.Debug("[push-redis] tasks are down")

	// delete service
	service, err := p.deleteService(instance, describedService)
	if err != nil {
		ch <- &deprovisionPushRedisResult{err: err}
		return
	}
	p.logger.Debug("[push-redis] did delete service")

	// wait service to stop
	waitServiceCh := make(chan bool)
	go waitServiceDown(p.logger, instance, waitServiceCh, p.describeService)
	if serviceDown := <-waitServiceCh; !serviceDown {
		ch <- &deprovisionPushRedisResult{err: errors.New("[push-redis] service did not go down")}
		return
	}
	p.logger.Debug("[push-redis] service is down")

	// delete service discovery
	serviceDiscovery, err := p.deleteServiceDiscovery(instance)
	if err != nil {
		ch <- &deprovisionPushRedisResult{err: err}
		return
	}
	p.logger.Debug("[push-redis] did delete service discovery")

	ch <- &deprovisionPushRedisResult{
		service:          service,
		serviceDiscovery: serviceDiscovery,
	}
}

// TODO refactor
func (p *ecsPushRedisProvisioner) stopService(instance *models.Instance, describeService *ecs.DescribeServicesOutput) (*ecs.UpdateServiceOutput, error) {
	return p.provisionerConfig.ecs.UpdateService(&ecs.UpdateServiceInput{
		Cluster:      p.provisionerConfig.cluster,
		DesiredCount: aws.Int64(0),
		Service:      describeService.Services[0].ServiceName,
	})
}

func (p *ecsPushRedisProvisioner) deleteService(instance *models.Instance, describeService *ecs.DescribeServicesOutput) (*ecs.DeleteServiceOutput, error) {
	return p.provisionerConfig.ecs.DeleteService(&ecs.DeleteServiceInput{
		Cluster: p.provisionerConfig.cluster,
		Force: aws.Bool(true),
		Service: describeService.Services[0].ServiceName,
	})
}

func (p *ecsPushRedisProvisioner) deleteServiceDiscovery(instance *models.Instance) (*servicediscovery.DeleteServiceOutput, error) {
	listServiceResult, err := listServiceDiscoveryServices(p.provisionerConfig.serviceDiscovery)
	if err != nil {
		return nil, nil
	}

	for _, service := range listServiceResult.Services {
		if *service.Name == pushRedisWithInstance(instance.Name) {
			return p.provisionerConfig.serviceDiscovery.DeleteService(&servicediscovery.DeleteServiceInput{
				Id: service.Id,
			})
		}
	}

	return nil, errors.New(fmt.Sprintf("could not find push-redis service discovery service for instance %s", instance.Name))
}

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
