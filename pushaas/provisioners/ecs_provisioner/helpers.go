package ecs_provisioner

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"go.uber.org/zap"

	"github.com/pushaas/pushaas/pushaas/models"
)

/*
	===========================================================================
	iam
	===========================================================================
*/
const roleName = "ecsTaskExecutionRole"

func getIamRole(iamSvc iamiface.IAMAPI) (*iam.GetRoleOutput, error) {
	input := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}

	return iamSvc.GetRole(input)
}

/*
	===========================================================================
	ecs
	===========================================================================
*/
func describeService(instanceName string, provisionerConfig *EcsProvisionerConfig) (*ecs.DescribeServicesOutput, error) {
	return provisionerConfig.ecs.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  provisionerConfig.cluster,
		Services: []*string{aws.String(instanceName)},
	})
}

func deleteService(describeServiceOutput *ecs.DescribeServicesOutput, provisionerConfig *EcsProvisionerConfig) (*ecs.DeleteServiceOutput, error) {
	return provisionerConfig.ecs.DeleteService(&ecs.DeleteServiceInput{
		Cluster: provisionerConfig.cluster,
		Force: aws.Bool(true),
		Service: describeServiceOutput.Services[0].ServiceName,
	})
}

func stopService(describeService *ecs.DescribeServicesOutput, provisionerConfig *EcsProvisionerConfig) (*ecs.UpdateServiceOutput, error) {
	return provisionerConfig.ecs.UpdateService(&ecs.UpdateServiceInput{
		Cluster:      provisionerConfig.cluster,
		DesiredCount: aws.Int64(0),
		Service:      describeService.Services[0].ServiceName,
	})
}

func deleteTaskDefinition(describeService *ecs.DescribeServicesOutput, provisionerConfig *EcsProvisionerConfig) (*ecs.DeregisterTaskDefinitionOutput, error) {
	return provisionerConfig.ecs.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: describeService.Services[0].TaskDefinition,
	})
}

/*
	===========================================================================
	serviceDiscovery
	===========================================================================
*/
func describeNamespace(provisionerConfig *EcsProvisionerConfig) (*servicediscovery.GetNamespaceOutput, error) {
	return provisionerConfig.serviceDiscovery.GetNamespace(&servicediscovery.GetNamespaceInput{
		Id: provisionerConfig.dnsNamespace,
	})
}

func listServiceDiscoveryInstances(instanceName string, provisionerConfig *EcsProvisionerConfig) (*servicediscovery.DiscoverInstancesOutput, error) {
	namespaceOutput, err := describeNamespace(provisionerConfig)
	if err != nil {
		return nil, err
	}

	return provisionerConfig.serviceDiscovery.DiscoverInstances(&servicediscovery.DiscoverInstancesInput{
		NamespaceName: namespaceOutput.Namespace.Name,
		ServiceName: aws.String(instanceName),
	})
}

func deleteServiceDiscoveryInstances(instanceName string, provisionerConfig *EcsProvisionerConfig) (*servicediscovery.DeregisterInstanceOutput, error) {
	instancesOutput, err := listServiceDiscoveryInstances(instanceName, provisionerConfig)
	if err != nil {
		return nil, nil
	}

	// we are assuming here there is always a single instance
	serviceDiscoveryInstance := instancesOutput.Instances[0]
	return provisionerConfig.serviceDiscovery.DeregisterInstance(&servicediscovery.DeregisterInstanceInput{
		InstanceId: serviceDiscoveryInstance.InstanceId,
	})
}

func listServiceDiscoveryServices(provisionerConfig *EcsProvisionerConfig) (*servicediscovery.ListServicesOutput, error) {
	return provisionerConfig.serviceDiscovery.ListServices(&servicediscovery.ListServicesInput{})
}

func deleteServiceDiscovery(instanceName string, provisionerConfig *EcsProvisionerConfig) (*servicediscovery.DeleteServiceOutput, error) {
	listServiceResult, err := listServiceDiscoveryServices(provisionerConfig)
	if err != nil {
		return nil, nil
	}

	for _, service := range listServiceResult.Services {
		if *service.Name == instanceName {
			return provisionerConfig.serviceDiscovery.DeleteService(&servicediscovery.DeleteServiceInput{
				Id: service.Id,
			})
		}
	}

	return nil, errors.New(fmt.Sprintf("could not find service discovery service for instance %s", instanceName))
}


/*
	===========================================================================
	other
	===========================================================================
*/
const attempts = 60
const interval = 5 * time.Second

func waitTrue(ch chan bool, evaluationFn func(attempt int) bool) {
	for i := 0; i < attempts; i++ {
		isLastAttempt := i+1 == attempts
		isSuccess := evaluationFn(i)

		if !isSuccess {
			if isLastAttempt {
				ch <- false
				return
			}
			time.Sleep(interval)
			continue
		}
		ch <- true
		return
	}
}

func waitServiceUp(logger *zap.Logger, instance *models.Instance, ch chan bool, describeServiceFunc func(*models.Instance) (*ecs.DescribeServicesOutput, error)) {
	waitTrue(ch, func(attempt int) bool {
		serviceResult, err := describeServiceFunc(instance)
		if err != nil {
			logger.Error(fmt.Sprintf("[waitServiceUp] failed on attempt %d", attempt), zap.Error(err))
			return false
		}
		isServiceUp := len(serviceResult.Services) > 0 && *serviceResult.Services[0].RunningCount > 0
		logger.Debug(fmt.Sprintf("[waitServiceUp] attempt %d with result isServiceUp=%t", attempt, isServiceUp), zap.Error(err))
		return isServiceUp
	})
}

func waitServiceStopAllTasks(logger *zap.Logger, instance *models.Instance, ch chan bool, describeServiceFunc func(*models.Instance) (*ecs.DescribeServicesOutput, error)) {
	waitTrue(ch, func(attempt int) bool {
		serviceResult, err := describeServiceFunc(instance)
		if err != nil {
			logger.Error(fmt.Sprintf("[waitServiceStopAllTasks] failed on attempt %d", attempt), zap.Error(err))
		}
		areTasksStoped := *serviceResult.Services[0].RunningCount == 0
		logger.Debug(fmt.Sprintf("[waitServiceStopAllTasks] attempt %d with result areTasksStoped=%t", attempt, areTasksStoped), zap.Error(err))
		return areTasksStoped
	})
}
func waitServiceDown(logger *zap.Logger, instance *models.Instance, ch chan bool, describeServiceFunc func(*models.Instance) (*ecs.DescribeServicesOutput, error)) {
	waitTrue(ch, func(attempt int) bool {
		serviceResult, err := describeServiceFunc(instance)
		if err != nil {
			logger.Error(fmt.Sprintf("[waitServiceDown] failed on attempt %d", attempt), zap.Error(err))
		}
		noServices := len(serviceResult.Services) == 0
		if noServices {
			logger.Debug(fmt.Sprintf("[waitServiceDown] attempt %d with result isServiceDown=true (len services=%d)", attempt, len(serviceResult.Services)), zap.Error(err))
			return true
		}

		isServiceDown := true
		for _, service := range serviceResult.Services {
			// there is no constant for this "INACTIVE" ¯\_(ツ)_/¯
			if *service.Status != "INACTIVE" {
				isServiceDown = false
				break
			}
		}

		logger.Debug(fmt.Sprintf("[waitServiceDown] attempt %d with result isServiceDown=%t (len services=%d)", attempt, isServiceDown, len(serviceResult.Services)), zap.Error(err))
		return isServiceDown
	})
}

// TODO technical debt
func waitTaskNetworkInterface(logger *zap.Logger, instance *models.Instance, ch chan bool, describeEniFunc func(instance *models.Instance) (*ec2.DescribeNetworkInterfacesOutput, error)) {
	waitTrue(ch, func(attempt int) bool {
		eni, err := describeEniFunc(instance)
		if err != nil {
			logger.Error(fmt.Sprintf("[waitTaskNetworkInterface] failed on attempt %d", attempt), zap.Error(err))
		}
		isEniUp := len(eni.NetworkInterfaces) > 0 && eni.NetworkInterfaces[0].Association != nil && *eni.NetworkInterfaces[0].Association.PublicIp != ""
		logger.Debug(fmt.Sprintf("[waitTaskNetworkInterface] attempt %d with result isEniUp=%t", attempt, isEniUp), zap.Error(err))
		return isEniUp
	})
}
