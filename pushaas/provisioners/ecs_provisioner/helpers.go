package ecs_provisioner

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicediscovery/servicediscoveryiface"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/models"
)

const roleName = "ecsTaskExecutionRole"

func getIamRole(iamSvc iamiface.IAMAPI) (*iam.GetRoleOutput, error) {
	input := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}

	return iamSvc.GetRole(input)
}

func listServiceDiscoveryServices(serviceDiscovery servicediscoveryiface.ServiceDiscoveryAPI) (*servicediscovery.ListServicesOutput, error) {
	return serviceDiscovery.ListServices(&servicediscovery.ListServicesInput{})
}

const attempts = 30
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
