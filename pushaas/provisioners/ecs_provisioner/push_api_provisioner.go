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
		Deprovision(*models.Instance, chan *deprovisionPushApiResult)
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

	deprovisionPushApiResult struct {
		service          *ecs.DeleteServiceOutput
		serviceDiscovery *servicediscovery.DeleteServiceOutput
		taskDefinition   *ecs.DeregisterTaskDefinitionOutput
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
	taskDefinition, err := p.createTaskDefinition(instance, role, eni)
	if err != nil {
		ch <- &provisionPushApiResult{err: err}
		return
	}
	p.logger.Debug("[push-api] did create task definition")

	// create service discovery
	serviceDiscovery, err := p.createServiceDiscovery(instance)
	if err != nil {
		ch <- &provisionPushApiResult{err: err}
		return
	}
	p.logger.Debug("[push-api] did create service discovery")

	// create service
	service, err := p.createService(instance, serviceDiscovery)
	if err != nil {
		ch <- &provisionPushApiResult{err: err}
		return
	}
	p.logger.Debug("[push-api] did create service")

	// wait for service to go up
	waitCh := make(chan bool)
	go waitServiceUp(p.logger, instance, waitCh, p.describeService)
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

func (p *ecsPushApiProvisioner) createTaskDefinition(
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
						Name:  aws.String("PUSHAPI_PUSH_STREAM__URL"),
						Value: aws.String(fmt.Sprintf("http://%s.tsuru:9080", pushStreamPublicIp)),
					},
				},
			},
		},
	})
}

func (p *ecsPushApiProvisioner) createServiceDiscovery(instance *models.Instance) (*servicediscovery.CreateServiceOutput, error) {
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

func (p *ecsPushApiProvisioner) createService(instance *models.Instance, serviceDiscovery *servicediscovery.CreateServiceOutput) (*ecs.CreateServiceOutput, error) {
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
func (p *ecsPushApiProvisioner) Deprovision(instance *models.Instance, ch chan *deprovisionPushApiResult) {
	var err error

	// get service
	describedService, err := p.describeService(instance)
	if err != nil {
		ch <- &deprovisionPushApiResult{err: err}
		return
	}
	if len(describedService.Services) == 0 {
		ch <- &deprovisionPushApiResult{err: errors.New(fmt.Sprintf("[push-api] could  not find service %s", pushApiWithInstance(instance.Name)))}
		return
	}
	p.logger.Debug("[push-api] did locate service")

	// scale to 0 tasks
	_, err = p.stopService(instance, describedService)
	if err != nil {
		ch <- &deprovisionPushApiResult{err: err}
		return
	}
	p.logger.Debug("[push-api] did update service to desiredCount 0")

	// wait tasks to stop
	waitTasksCh := make(chan bool)
	go waitServiceStopAllTasks(p.logger, instance, waitTasksCh, p.describeService)
	if serviceDown := <-waitTasksCh; !serviceDown {
		ch <- &deprovisionPushApiResult{err: errors.New("[push-api] service did not remove all tasks")}
		return
	}
	p.logger.Debug("[push-api] tasks are down")

	// delete service
	service, err := p.deleteService(instance, describedService)
	if err != nil {
		ch <- &deprovisionPushApiResult{err: err}
		return
	}
	p.logger.Debug("[push-api] did delete service")

	// wait service to stop
	waitServiceCh := make(chan bool)
	go waitServiceDown(p.logger, instance, waitServiceCh, p.describeService)
	if serviceDown := <-waitServiceCh; !serviceDown {
		ch <- &deprovisionPushApiResult{err: errors.New("[push-api] service did not go down")}
		return
	}
	p.logger.Debug("[push-api] service is down")

	// delete service discovery
	serviceDiscovery, err := p.deleteServiceDiscovery(instance)
	if err != nil {
		ch <- &deprovisionPushApiResult{err: err}
		return
	}
	p.logger.Debug("[push-api] did delete service discovery")

	// delete task definition
	taskDefinition, err := p.deleteTaskDefinition(instance, describedService)
	if err != nil {
		ch <- &deprovisionPushApiResult{err: err}
		return
	}
	p.logger.Debug("[push-api] did delete task definition")

	ch <- &deprovisionPushApiResult{
		service:          service,
		serviceDiscovery: serviceDiscovery,
		taskDefinition:   taskDefinition,
	}
}

// TODO refactor
func (p *ecsPushApiProvisioner) stopService(instance *models.Instance, describeService *ecs.DescribeServicesOutput) (*ecs.UpdateServiceOutput, error) {
	return p.provisionerConfig.ecs.UpdateService(&ecs.UpdateServiceInput{
		Cluster:      p.provisionerConfig.cluster,
		DesiredCount: aws.Int64(0),
		Service:      describeService.Services[0].ServiceName,
	})
}

func (p *ecsPushApiProvisioner) deleteService(instance *models.Instance, describeService *ecs.DescribeServicesOutput) (*ecs.DeleteServiceOutput, error) {
	return p.provisionerConfig.ecs.DeleteService(&ecs.DeleteServiceInput{
		Cluster: p.provisionerConfig.cluster,
		Force:   aws.Bool(true),
		Service: describeService.Services[0].ServiceName,
	})
}

func (p *ecsPushApiProvisioner) deleteServiceDiscovery(instance *models.Instance) (*servicediscovery.DeleteServiceOutput, error) {
	listServiceResult, err := listServiceDiscoveryServices(p.provisionerConfig.serviceDiscovery)
	if err != nil {
		return nil, nil
	}

	for _, service := range listServiceResult.Services {
		if *service.Name == pushApiWithInstance(instance.Name) {
			return p.provisionerConfig.serviceDiscovery.DeleteService(&servicediscovery.DeleteServiceInput{
				Id: service.Id,
			})
		}
	}

	return nil, errors.New(fmt.Sprintf("could not find push-api service discovery service for instance %s", instance.Name))
}

func (p *ecsPushApiProvisioner) deleteTaskDefinition(instance *models.Instance, describeService *ecs.DescribeServicesOutput) (*ecs.DeregisterTaskDefinitionOutput, error) {
	return p.provisionerConfig.ecs.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: describeService.Services[0].TaskDefinition,
	})
}

/*
	===========================================================================
	other
	===========================================================================
*/
func (p *ecsPushApiProvisioner) describeService(instance *models.Instance) (*ecs.DescribeServicesOutput, error) {
	return p.provisionerConfig.ecs.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  p.provisionerConfig.cluster,
		Services: []*string{aws.String(pushApiWithInstance(instance.Name))},
	})
}

func NewEcsPushApiProvisioner(logger *zap.Logger, provisionerConfig *EcsProvisionerConfig) EcsPushApiProvisioner {
	return &ecsPushApiProvisioner{
		logger:            logger,
		provisionerConfig: provisionerConfig,
	}
}
