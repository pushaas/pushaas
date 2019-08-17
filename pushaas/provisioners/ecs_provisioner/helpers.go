package ecs_provisioner

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/rafaeleyng/pushaas/pushaas/models"
)

const roleName = "ecsTaskExecutionRole"

func getIamRole(iamSvc iamiface.IAMAPI) (*iam.GetRoleOutput, error) {
	input := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}

	return iamSvc.GetRole(input)
}

const attempts = 30
const interval = 5 * time.Second

func waitSuccess(ch chan bool, evaluationFn func() bool) {
	for i := 0; i < attempts; i++ {
		isLastAttempt := i+1 == attempts
		isSuccess := evaluationFn()

		if !isSuccess {
			if isLastAttempt {
				ch <- false
				break
			}
			time.Sleep(interval)
			continue
		}
		ch <- true
	}
}

func waitServiceUp(instance *models.Instance, ch chan bool, describeServiceFunc func(*models.Instance) (*ecs.DescribeServicesOutput, error)) {
	waitSuccess(ch, func() bool {
		serviceResult, err := describeServiceFunc(instance)
		isServiceUp := err == nil && len(serviceResult.Services) > 0 && *serviceResult.Services[0].RunningCount > 0
		return isServiceUp
	})
}

func waitServiceDown(instance *models.Instance, ch chan bool, describeServiceFunc func(*models.Instance) (*ecs.DescribeServicesOutput, error)) {
	waitSuccess(ch, func() bool {
		serviceResult, err := describeServiceFunc(instance)
		isServiceDown := err == nil || len(serviceResult.Services) == 0 || *serviceResult.Services[0].RunningCount == 0
		return !isServiceDown
	})
}

// TODO technical debt
func waitTaskNetworkInterface(instance *models.Instance, ch chan bool, describeEniFunc func(instance *models.Instance) (*ec2.DescribeNetworkInterfacesOutput, error)) {
	waitSuccess(ch, func() bool {
		eni, err := describeEniFunc(instance)
		isEniUp := err == nil && len(eni.NetworkInterfaces) > 0 && eni.NetworkInterfaces[0].Association != nil && eni.NetworkInterfaces[0].Association.PublicIp != nil
		return isEniUp
	})
}
