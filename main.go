package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"

	//"github.com/rafaeleyng/pushaas/pushaas"
)

const roleName = "ecsTaskExecutionRole"

const instanceName = "instance-123"

const clusterName = "pushaas-cluster"
const redisTaskDefinitionName = "push-redis-" + instanceName
const serviceName = "service-redis-" + instanceName

func getIamRole(iamSvc *iam.IAM) {
	input := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}

	output, err := iamSvc.GetRole(input)
	if err != nil {
		fmt.Println("========== FAILED GetRole ==========")
		panic(err)
	}
	fmt.Println("========== GetRole ==========")
	fmt.Println(output.GoString())
}

func listTaskDescriptions(svc *ecs.ECS) {
	input := &ecs.ListTaskDefinitionsInput{}

	output, err := svc.ListTaskDefinitions(input)
	if err != nil {
		fmt.Println("========== FAILED ListTaskDefinitions ==========")
		panic(err)
	}
	fmt.Println("========== ListTaskDefinitions ==========")
	fmt.Println(output.GoString())
}

//func createRedisTaskDefinition(svc *ecs.ECS) {
//	input := &ecs.RegisterTaskDefinitionInput{
//
//	}
//
//	svc.RegisterTaskDefinition()
//}

func listServices(svc *ecs.ECS) {
	input := &ecs.ListServicesInput{
		Cluster: aws.String(clusterName),
		LaunchType: aws.String(ecs.LaunchTypeFargate),
	}

	output, err := svc.ListServices(input)
	if err != nil {
		fmt.Println("========== FAILED ListServices ==========")
		panic(err)
	}
	fmt.Println("========== ListServices ==========")
	fmt.Println(output.GoString())
}

func describeRedisService(svc *ecs.ECS) {
	input := &ecs.DescribeServicesInput{
		Cluster: aws.String(clusterName),
		Services: []*string{aws.String(serviceName)},
	}

	output, err := svc.DescribeServices(input)
	if err != nil {
		fmt.Println("========== FAILED DescribeServices ==========")
		panic(err)
	}
	fmt.Println("========== DescribeServices ==========")
	fmt.Println(output.GoString())
}

func createRedisService(svc *ecs.ECS) {
	input := &ecs.CreateServiceInput{
		Cluster: aws.String(clusterName),
		DesiredCount: aws.Int64(1),
		ServiceName: aws.String(serviceName),
		TaskDefinition: aws.String(redisTaskDefinitionName),
		LaunchType: aws.String(ecs.LaunchTypeFargate),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
				// TODO is comming from `scripts/40-pushaas/60-create-app-service/terraform.tfstate`, should create specific for redis services
				SecurityGroups: []*string{aws.String("sg-0ab3dd2d1b0b21f1b")},
				// TODO is comming from `scripts/10-vpc/10-create-vpc/terraform.tfstate`, this is ok, just pass as env
				Subnets: []*string{aws.String("subnet-0852fc9806179665c")},
			},
		},
	}

	output, err := svc.CreateService(input)
	if err != nil {
		fmt.Println("========== FAILED CreateService ==========")
		panic(err)
	}
	fmt.Println("========== CreateService ==========")
	fmt.Println(output.GoString())
}

func deleteRedisService(svc *ecs.ECS) {
	input := &ecs.DeleteServiceInput{
		Cluster: aws.String(clusterName),
		Force: aws.Bool(true),
		Service: aws.String(serviceName),
	}

	output, err := svc.DeleteService(input)
	if err != nil {
		fmt.Println("========== FAILED DeleteService ==========")
		panic(err)
	}
	fmt.Println("========== DeleteService ==========")
	fmt.Println(output.GoString())
}

func main() {
	//pushaas.Run()

	mySession := session.Must(session.NewSession())
	ecsSvc := ecs.New(mySession)
	iamSvc := iam.New(mySession)

	getIamRole(iamSvc)

	listTaskDescriptions(ecsSvc)

	//listServices(ecsSvc)
	//describeRedisService(ecsSvc)
	//deleteRedisService(ecsSvc)
	//createRedisService(ecsSvc)
}
