package main

//import "github.com/rafaeleyng/pushaas/pushaas"

import (
	"github.com/rafaeleyng/pushaas/pushaas"
)

func main() {
	pushaas.Run()

	//mySession := session.Must(session.NewSession())
	////if err != nil {
	////	fmt.Println(err)
	////	return
	////}
	//
	//clusterName := aws.String("pushaas-cluster")
	//svc := ecs.New(mySession)

	//input := &ecs.ListServicesInput{
	//	Cluster: clusterName,
	//}
	//input.SetLaunchType(ecs.LaunchTypeFargate)
	//
	//output, err := svc.ListServices(input)
	//if err != nil {
	//	fmt.Println("### 1")
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(output.GoString())

	// -----------------------------------------------

	//serviceName := aws.String("apagar-1")
	//taskDefinition := aws.String("pushaas-app-task")
	//desiredCount := aws.Int64(1)
	//assignPublicIp := aws.String(ecs.AssignPublicIpEnabled)
	//launchType := aws.String(ecs.LaunchTypeFargate)
	//securityGroup := aws.String("sg-04c7bdab8676d92a3")
	//subnet := aws.String("subnet-0852fc9806179665c")
	//
	//createServiceInput := &ecs.CreateServiceInput{
	//	Cluster: clusterName,
	//	DesiredCount: desiredCount,
	//	ServiceName: serviceName,
	//	TaskDefinition: taskDefinition,
	//	LaunchType: launchType,
	//	NetworkConfiguration: &ecs.NetworkConfiguration{
	//		AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
	//			AssignPublicIp: assignPublicIp,
	//			SecurityGroups: []*string{securityGroup},
	//			Subnets: []*string{subnet},
	//		},
	//	},
	//}
	//
	//_, err := svc.CreateService(createServiceInput)
	//if err != nil {
	//	fmt.Println("### 2")
	//	fmt.Println(err)
	//	return
	//}

	// -----------------------------------------------

	//input := &ecs.DescribeServicesInput{
	//	Cluster: clusterName,
	//	Services: []*string{
	//		aws.String("apagar-2"),
	//	},
	//}
	//
	//result, err := svc.DescribeServices(input)
	//if err != nil {
	//	if aerr, ok := err.(awserr.Error); ok {
	//			fmt.Println(aerr.Error())
	//	} else {
	//		fmt.Println(err.Error())
	//	}
	//	return
	//}
	//
	////fmt.Println(result.Services[0].TaskSets)
	//fmt.Println(result)

	// -----------------------------------------------

	//input := &ecs.DescribeTasksInput{
	//	//Clus
	//	Tasks: []*string{
	//		aws.String("c5cba4eb-5dad-405e-96db-71ef8eefe6a8"),
	//	},
	//}
	//
	//result, err := svc.DescribeTasks(input)
	//if err != nil {
	//	if aerr, ok := err.(awserr.Error); ok {
	//		switch aerr.Code() {
	//		case ecs.ErrCodeServerException:
	//			fmt.Println(ecs.ErrCodeServerException, aerr.Error())
	//		case ecs.ErrCodeClientException:
	//			fmt.Println(ecs.ErrCodeClientException, aerr.Error())
	//		case ecs.ErrCodeInvalidParameterException:
	//			fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
	//		case ecs.ErrCodeClusterNotFoundException:
	//			fmt.Println(ecs.ErrCodeClusterNotFoundException, aerr.Error())
	//		default:
	//			fmt.Println(aerr.Error())
	//		}
	//	} else {
	//		// Print the error, cast err to awserr.Error to get the Code and
	//		// Message from an error.
	//		fmt.Println(err.Error())
	//	}
	//	return
	//}
	//
	//fmt.Println(result)
}
