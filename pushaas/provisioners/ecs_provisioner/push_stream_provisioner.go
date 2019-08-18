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
		Deprovision(*models.Instance, chan *deprovisionPushStreamResult)
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

	deprovisionPushStreamResult struct {
		service          *ecs.DeleteServiceOutput
		serviceDiscovery *servicediscovery.DeleteServiceOutput
		taskDefinition   *ecs.DeregisterTaskDefinitionOutput
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
	taskDefinition, err := p.createTaskDefinition(instance, role)
	if err != nil {
		ch <- &provisionPushStreamResult{err: err}
		return
	}
	p.logger.Debug("[push-stream] did create task definition")

	// create service discovery
	serviceDiscovery, err := p.createServiceDiscovery(instance)
	if err != nil {
		ch <- &provisionPushStreamResult{err: err}
		return
	}
	p.logger.Debug("[push-stream] did create service discovery")

	// create service
	service, err := p.createService(instance, serviceDiscovery)
	if err != nil {
		ch <- &provisionPushStreamResult{err: err}
		return
	}
	p.logger.Debug("[push-stream] did create service")

	// wait for service to go up
	waitCh := make(chan bool)
	go waitServiceUp(p.logger, instance, waitCh, p.describeService)
	if serviceUp := <-waitCh; !serviceUp {
		ch <- &provisionPushStreamResult{err: errors.New("push-stream service did not become available")}
		return
	}
	p.logger.Debug("[push-stream] service is up")

	// wait for network interface
	eniCh := make(chan bool)
	go waitTaskNetworkInterface(p.logger, instance, eniCh, p.describeTaskNetworkInterface)
	if isEniUp := <-eniCh; !isEniUp {
		ch <- &provisionPushStreamResult{err: errors.New("push-stream ENI failed to become available")}
		return
	}

	// get network interface
	eni, err := p.describeTaskNetworkInterface(instance)
	if err != nil {
		ch <- &provisionPushStreamResult{err: err}
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

func (p *ecsPushStreamProvisioner) createTaskDefinition(instance *models.Instance, role *iam.GetRoleOutput) (*ecs.RegisterTaskDefinitionOutput, error) {
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

func (p *ecsPushStreamProvisioner) createServiceDiscovery(instance *models.Instance) (*servicediscovery.CreateServiceOutput, error) {
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

func (p *ecsPushStreamProvisioner) createService(instance *models.Instance, pushStreamDiscovery *servicediscovery.CreateServiceOutput) (*ecs.CreateServiceOutput, error) {
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
func (p *ecsPushStreamProvisioner) Deprovision(instance *models.Instance, ch chan *deprovisionPushStreamResult) {
	var err error

	// get service
	describedService, err := p.describeService(instance)
	if err != nil {
		ch <- &deprovisionPushStreamResult{err: err}
		return
	}
	if len(describedService.Services) == 0 {
		ch <- &deprovisionPushStreamResult{err: errors.New(fmt.Sprintf("[push-stream] could  not find service %s", pushStreamWithInstance(instance.Name)))}
		return
	}
	p.logger.Debug("[push-stream] did locate service")

	// scale to 0 tasks
	_, err = p.stopService(instance, describedService)
	if err != nil {
		ch <- &deprovisionPushStreamResult{err: err}
		return
	}
	p.logger.Debug("[push-stream] did update service to desiredCount 0")

	// wait tasks to stop
	waitTasksCh := make(chan bool)
	go waitServiceStopAllTasks(p.logger, instance, waitTasksCh, p.describeService)
	if serviceDown := <-waitTasksCh; !serviceDown {
		ch <- &deprovisionPushStreamResult{err: errors.New("[push-stream] service did not remove all tasks")}
		return
	}
	p.logger.Debug("[push-stream] service is down")

	// delete service
	service, err := p.deleteService(instance, describedService)
	if err != nil {
		ch <- &deprovisionPushStreamResult{err: err}
		return
	}
	p.logger.Debug("[push-stream] did delete service")

	// wait service to stop
	waitServiceCh := make(chan bool)
	go waitServiceDown(p.logger, instance, waitServiceCh, p.describeService)
	if serviceDown := <-waitServiceCh; !serviceDown {
		ch <- &deprovisionPushStreamResult{err: errors.New("[push-service] service did not go down")}
		return
	}
	p.logger.Debug("[push-stream] service is down")

	// delete service discovery
	serviceDiscovery, err := p.deleteServiceDiscovery(instance)
	if err != nil {
		ch <- &deprovisionPushStreamResult{err: err}
		return
	}
	p.logger.Debug("[push-stream] did delete service discovery")

	// delete task definition
	taskDefinition, err := p.deleteTaskDefinition(instance, describedService)
	if err != nil {
		ch <- &deprovisionPushStreamResult{err: err}
		return
	}
	p.logger.Debug("[push-stream] did delete task definition")

	ch <- &deprovisionPushStreamResult{
		service:          service,
		serviceDiscovery: serviceDiscovery,
		taskDefinition:   taskDefinition,
	}
}

// TODO refactor
func (p *ecsPushStreamProvisioner) stopService(instance *models.Instance, describeService *ecs.DescribeServicesOutput) (*ecs.UpdateServiceOutput, error) {
	return p.provisionerConfig.ecs.UpdateService(&ecs.UpdateServiceInput{
		Cluster:      p.provisionerConfig.cluster,
		DesiredCount: aws.Int64(0),
		Service:      describeService.Services[0].ServiceName,
	})
}

func (p *ecsPushStreamProvisioner) deleteService(instance *models.Instance, describeService *ecs.DescribeServicesOutput) (*ecs.DeleteServiceOutput, error) {
	return p.provisionerConfig.ecs.DeleteService(&ecs.DeleteServiceInput{
		Cluster: p.provisionerConfig.cluster,
		Force: aws.Bool(true),
		Service: describeService.Services[0].ServiceName,
	})
}

func (p *ecsPushStreamProvisioner) deleteServiceDiscovery(instance *models.Instance) (*servicediscovery.DeleteServiceOutput, error) {
	listServiceResult, err := listServiceDiscoveryServices(p.provisionerConfig.serviceDiscovery)
	if err != nil {
		return nil, nil
	}

	for _, service := range listServiceResult.Services {
		if *service.Name == pushStreamWithInstance(instance.Name) {
			return p.provisionerConfig.serviceDiscovery.DeleteService(&servicediscovery.DeleteServiceInput{
				Id: service.Id,
			})
		}
	}

	return nil, errors.New(fmt.Sprintf("could not find push-stream service discovery service for instance %s", instance.Name))
}

func (p *ecsPushStreamProvisioner) deleteTaskDefinition(instance *models.Instance, describeService *ecs.DescribeServicesOutput) (*ecs.DeregisterTaskDefinitionOutput, error) {
	return p.provisionerConfig.ecs.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: describeService.Services[0].TaskDefinition,
	})
}

/*
	===========================================================================
	other
	===========================================================================
*/
func (p *ecsPushStreamProvisioner) listTasks(instance *models.Instance) (*ecs.ListTasksOutput, error) {
	return p.provisionerConfig.ecs.ListTasks(&ecs.ListTasksInput{
		Cluster:     p.provisionerConfig.cluster,
		ServiceName: aws.String(pushStreamWithInstance(instance.Name)),
	})
}

func (p *ecsPushStreamProvisioner) describeTasks(instance *models.Instance) (*ecs.DescribeTasksOutput, error) {
	listOutput, err := p.listTasks(instance)
	if err != nil {
		return nil, err
	}

	if len(listOutput.TaskArns) == 0 {
		return nil, errors.New(fmt.Sprintf("[describeTasks] no tasks in service %s", pushStreamWithInstance(instance.Name)))
	}

	return p.provisionerConfig.ecs.DescribeTasks(&ecs.DescribeTasksInput{
		Tasks:   []*string{listOutput.TaskArns[0]},
		Cluster: p.provisionerConfig.cluster,
	})
}

func (p *ecsPushStreamProvisioner) describeTaskNetworkInterface(instance *models.Instance) (*ec2.DescribeNetworkInterfacesOutput, error) {
	describeOutput, err := p.describeTasks(instance)
	if err != nil {
		return nil, err
	}

	if len(describeOutput.Tasks) == 0 || len(describeOutput.Tasks[0].Attachments) == 0 {
		return nil, errors.New(fmt.Sprintf("[describeTaskNetworkInterface] no tasks or attachments found for service %s", pushStreamWithInstance(instance.Name)))
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
