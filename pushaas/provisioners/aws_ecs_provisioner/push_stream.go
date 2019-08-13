package aws_ecs_provisioner

const pushAgent = "push-agent"
const pushStream = "push-stream"

func pushStreamWithInstance(instanceName string) string {
	return pushStream + "-" + instanceName
}

/////////////////////////////////////////////////////////////////////////////////
//// push-stream
/////////////////////////////////////////////////////////////////////////////////
//func createPushStreamTaskDefinition(svc *ecs.ECS, roleOutput *iam.GetRoleOutput) {
//	input := &ecs.RegisterTaskDefinitionInput{
//		Family: aws.String(pushStreamWithInstance),
//		ExecutionRoleArn: roleOutput.Role.Arn,
//		NetworkMode: aws.String(ecs.NetworkModeAwsvpc),
//		RequiresCompatibilities: []*string{aws.String(ecs.CompatibilityFargate)},
//		Cpu: aws.String("512"),
//		Memory: aws.String("1024"),
//		ContainerDefinitions: []*ecs.ContainerDefinition{
//			{
//				Cpu: aws.Int64(256),
//				Image: aws.String(pushStreamImage),
//				MemoryReservation: aws.Int64(512),
//				Name: aws.String(pushStream),
//				LogConfiguration: &ecs.LogConfiguration{
//					LogDriver: aws.String(ecs.LogDriverAwslogs),
//					Options: map[string]*string{
//						"awslogs-group": aws.String(logsGroup),
//						"awslogs-region": aws.String(awsRegion),
//						"awslogs-stream-prefix": aws.String(logsStreamPrefix),
//					},
//				},
//				PortMappings: []*ecs.PortMapping{
//					{
//						ContainerPort: aws.Int64(9080),
//						HostPort: aws.Int64(9080),
//					},
//				},
//			},
//			{
//				DependsOn: []*ecs.ContainerDependency{
//					{
//						Condition: aws.String(ecs.ContainerConditionStart),
//						ContainerName: aws.String(pushStream),
//					},
//				},
//				Cpu: aws.Int64(256),
//				Image: aws.String(pushAgentImage),
//				MemoryReservation: aws.Int64(512),
//				Name: aws.String(pushAgent),
//				LogConfiguration: &ecs.LogConfiguration{
//					LogDriver: aws.String(ecs.LogDriverAwslogs),
//					Options: map[string]*string{
//						"awslogs-group": aws.String(logsGroup),
//						"awslogs-region": aws.String(awsRegion),
//						"awslogs-stream-prefix": aws.String(logsStreamPrefix),
//					},
//				},
//				Environment: []*ecs.KeyValuePair{
//					{
//						Name: aws.String("PUSHAGENT_REDIS__URL"),
//						Value: aws.String("redis://" + pushRedisWithInstance + ".tsuru:6379"),
//					},
//					{
//						Name: aws.String("PUSHAGENT_PUSH_STREAM__URL"),
//						Value: aws.String("http://" + pushStreamWithInstance + ".tsuru:9080"),
//					},
//				},
//			},
//		},
//	}
//
//	output, err := svc.RegisterTaskDefinition(input)
//	if err != nil {
//		fmt.Println("========== push-stream - FAILED RegisterTaskDefinition ==========")
//		panic(err)
//	}
//	fmt.Println("========== push-stream - RegisterTaskDefinition ==========")
//	fmt.Println(output.GoString())
//}
//
//func describePushStreamService(svc *ecs.ECS) {
//	input := &ecs.DescribeServicesInput{
//		Cluster: aws.String(clusterName),
//		Services: []*string{aws.String(pushStreamWithInstance)},
//	}
//
//	output, err := svc.DescribeServices(input)
//	if err != nil {
//		fmt.Println("========== push-stream - FAILED DescribeServices ==========")
//		panic(err)
//	}
//	fmt.Println("========== push-stream - DescribeServices ==========")
//	fmt.Println(output.GoString())
//}
//
//func createPushStreamService(svc *ecs.ECS, pushStreamDiscovery *servicediscovery.CreateServiceOutput) {
//	input := &ecs.CreateServiceInput{
//		Cluster: aws.String(clusterName),
//		DesiredCount: aws.Int64(1),
//		ServiceName: aws.String(pushStreamWithInstance),
//		TaskDefinition: aws.String(pushStreamWithInstance),
//		LaunchType: aws.String(ecs.LaunchTypeFargate),
//		NetworkConfiguration: &ecs.NetworkConfiguration{
//			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
//				AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
//				SecurityGroups: []*string{aws.String(sg)},
//				Subnets: []*string{aws.String(subnet)},
//			},
//		},
//		ServiceRegistries: []*ecs.ServiceRegistry{
//			{
//				RegistryArn: pushStreamDiscovery.Service.Arn,
//			},
//		},
//	}
//
//	output, err := svc.CreateService(input)
//	if err != nil {
//		fmt.Println("========== push-stream - FAILED CreateService ==========")
//		panic(err)
//	}
//	fmt.Println("========== push-stream - CreateService ==========")
//	fmt.Println(output.GoString())
//}
//
//func deletePushStreamService(svc *ecs.ECS) {
//	input := &ecs.DeleteServiceInput{
//		Cluster: aws.String(clusterName),
//		Force: aws.Bool(true),
//		Service: aws.String(pushStreamWithInstance),
//	}
//
//	output, err := svc.DeleteService(input)
//	if err != nil {
//		fmt.Println("========== push-stream - FAILED DeleteService ==========")
//		panic(err)
//	}
//	fmt.Println("========== push-stream - DeleteService ==========")
//	fmt.Println(output.GoString())
//}
//
//func createPushStreamServiceDiscovery(svc *servicediscovery.ServiceDiscovery) *servicediscovery.CreateServiceOutput {
//	input := &servicediscovery.CreateServiceInput{
//		Name: aws.String(pushStreamWithInstance),
//		NamespaceId: aws.String(dnsNamespace),
//		DnsConfig: &servicediscovery.DnsConfig{
//			DnsRecords: []*servicediscovery.DnsRecord{
//				{
//					TTL: aws.Int64(10),
//					Type: aws.String("A"),
//				},
//			},
//		},
//		HealthCheckCustomConfig: &servicediscovery.HealthCheckCustomConfig{
//			FailureThreshold: aws.Int64(1),
//		},
//	}
//
//	output, err := svc.CreateService(input)
//	if err != nil {
//		fmt.Println("========== push-stream - FAILED CreateService discovery ==========")
//		panic(err)
//	}
//	fmt.Println("========== push-stream - CreateService discovery ==========")
//	fmt.Println(output.GoString())
//	return output
//}
//
//func listPushStreamTasks(svc *ecs.ECS) *ecs.ListTasksOutput {
//	input := &ecs.ListTasksInput{
//		Cluster: aws.String(clusterName),
//		ServiceName: aws.String(pushStreamWithInstance),
//	}
//
//	output, err := svc.ListTasks(input)
//	if err != nil {
//		fmt.Println("========== push-stream - FAILED ListTasks ==========")
//		panic(err)
//	}
//	fmt.Println("========== push-stream - ListTasks ==========")
//	fmt.Println(output.GoString())
//	return output
//}
//
//func describePushStreamTask(svc *ecs.ECS) *ecs.DescribeTasksOutput {
//	listOutput := listPushStreamTasks(svc)
//
//	input := &ecs.DescribeTasksInput{
//		Tasks: []*string{listOutput.TaskArns[0]},
//		Cluster: aws.String(clusterName),
//	}
//
//	output, err := svc.DescribeTasks(input)
//	if err != nil {
//		fmt.Println("========== push-stream - FAILED DescribeTasks ==========")
//		panic(err)
//	}
//	fmt.Println("========== push-stream - DescribeTasks ==========")
//	fmt.Println(output.GoString())
//	return output
//}
//
//func describePushStreamNetworkInterfaceTask(ecsSvc *ecs.ECS, ec2Svc *ec2.EC2) {
//	describeOutput := describePushStreamTask(ecsSvc)
//
//	var eniId *string
//	for _, kv := range describeOutput.Tasks[0].Attachments[0].Details {
//		if *kv.Name == "networkInterfaceId" {
//			eniId = kv.Value
//		}
//	}
//
//	input := &ec2.DescribeNetworkInterfacesInput{
//		NetworkInterfaceIds: []*string{eniId},
//	}
//
//	output, err := ec2Svc.DescribeNetworkInterfaces(input)
//	if err != nil {
//		fmt.Println("========== push-stream - FAILED DescribeNetworkInterfaces ==========")
//		panic(err)
//	}
//	fmt.Println("========== push-stream - DescribeNetworkInterfaces ==========")
//	fmt.Println(output.GoString())
//}
//
