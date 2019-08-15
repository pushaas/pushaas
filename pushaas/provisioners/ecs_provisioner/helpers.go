package ecs_provisioner

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

const roleName = "ecsTaskExecutionRole"

func getIamRole(iamSvc *iam.IAM) (*iam.GetRoleOutput, error) {
	input := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}

	return iamSvc.GetRole(input)
}

