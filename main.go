package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/servicediscovery"

	//"github.com/rafaeleyng/pushaas/pushaas"
)

/*
- [ok] service push-redis
- [ok] service discovery push-redis

- [ok] task definition push-api
- [ok] service push-api
- [ok] service discovery push-api

- [ok] task definition push-stream
- [ok] service push-stream
- [ok] service discovery push-stream
 */

const awsRegion = "us-east-1"

const logsGroup = "/ecs/pushaas"
const logsStreamPrefix = "ecs"

const roleName = "ecsTaskExecutionRole"

const instanceName = "instance-13"

const clusterName = "pushaas-cluster"

const pushRedis = "push-redis"
const pushRedisWithInstance = pushRedis + "-" + instanceName

const pushApiImage = "rafaeleyng/push-api:latest" // TODO use actual tag
const pushApi = "push-api"
const pushApiWithInstance = pushApi + "-" + instanceName

const pushAgentImage = "rafaeleyng/push-agent:latest" // TODO use actual tag
const pushAgent = "push-agent"

const pushStreamImage = "rafaeleyng/push-stream:latest" // TODO use actual tag
const pushStream = "push-stream"
const pushStreamWithInstance = pushStream + "-" + instanceName

// TODO comes from `scripts/40-pushaas/30-create-cluster/terraform.tfstate`, should create specific for each part of push service
const sgServiceInboudOutboundSubnet = "sg-0b5a8c5d666e24f25"
// TODO comes from `scripts/40-pushaas/60-create-app-service/terraform.tfstate`, should create specific for each part of push service
const sgServiceInboudAll = "sg-0aa587ddff427106d"
// TODO comes from `scripts/10-vpc/10-create-vpc/terraform.tfstate`, this is ok, just pass as env
const subnet = "subnet-0852fc9806179665c"
// TODO coms from `scripts/30-dns/10-create-namespace/terraform.tfstate`, this is ok, just pass as env
const dnsNamespace = "ns-srddhanacg4dxlea"

///////////////////////////////////////////////////////////////////////////////
// general
///////////////////////////////////////////////////////////////////////////////
func getIamRole(iamSvc *iam.IAM) *iam.GetRoleOutput {
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
	return output
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

///////////////////////////////////////////////////////////////////////////////
// redis
///////////////////////////////////////////////////////////////////////////////
func describeRedisService(svc *ecs.ECS) {
	input := &ecs.DescribeServicesInput{
		Cluster: aws.String(clusterName),
		Services: []*string{aws.String(pushRedisWithInstance)},
	}

	output, err := svc.DescribeServices(input)
	if err != nil {
		fmt.Println("========== redis - FAILED DescribeServices ==========")
		panic(err)
	}
	fmt.Println("========== redis - DescribeServices ==========")
	fmt.Println(output.GoString())
}

func createRedisService(svc *ecs.ECS, redisDiscovery *servicediscovery.CreateServiceOutput) {
	input := &ecs.CreateServiceInput{
		Cluster: aws.String(clusterName),
		DesiredCount: aws.Int64(1),
		ServiceName: aws.String(pushRedisWithInstance),
		TaskDefinition: aws.String(pushRedis),
		LaunchType: aws.String(ecs.LaunchTypeFargate),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
				SecurityGroups: []*string{aws.String(sgServiceInboudAll), aws.String(sgServiceInboudOutboundSubnet)},
				Subnets: []*string{aws.String(subnet)},
			},
		},
		ServiceRegistries: []*ecs.ServiceRegistry{
			{
				RegistryArn: redisDiscovery.Service.Arn,
			},
		},
	}

	output, err := svc.CreateService(input)
	if err != nil {
		fmt.Println("========== redis - FAILED CreateService ==========")
		panic(err)
	}
	fmt.Println("========== redis - CreateService ==========")
	fmt.Println(output.GoString())
}

func createRedisServiceDiscovery(svc *servicediscovery.ServiceDiscovery) *servicediscovery.CreateServiceOutput {
	input := &servicediscovery.CreateServiceInput{
		Name: aws.String(pushRedisWithInstance),
		NamespaceId: aws.String(dnsNamespace),
		DnsConfig: &servicediscovery.DnsConfig{
			DnsRecords: []*servicediscovery.DnsRecord{
				{
					TTL: aws.Int64(10),
					Type: aws.String("A"),
				},
			},
		},
		HealthCheckCustomConfig: &servicediscovery.HealthCheckCustomConfig{
			FailureThreshold: aws.Int64(1),
		},
	}

	output, err := svc.CreateService(input)
	if err != nil {
		fmt.Println("========== redis - FAILED CreateService discovery ==========")
		panic(err)
	}
	fmt.Println("========== redis - CreateService discovery ==========")
	fmt.Println(output.GoString())

	return output
}

func deleteRedisServiceDiscovery(svc *servicediscovery.ServiceDiscovery) {
	// TODO implement
}

func deleteRedisService(svc *ecs.ECS) {
	input := &ecs.DeleteServiceInput{
		Cluster: aws.String(clusterName),
		Force: aws.Bool(true),
		Service: aws.String(pushRedisWithInstance),
	}

	output, err := svc.DeleteService(input)
	if err != nil {
		fmt.Println("========== redis - FAILED DeleteService ==========")
		panic(err)
	}
	fmt.Println("========== redis - DeleteService ==========")
	fmt.Println(output.GoString())
}

///////////////////////////////////////////////////////////////////////////////
// push-api
///////////////////////////////////////////////////////////////////////////////
func createPushApiTaskDefinition(svc *ecs.ECS, roleOutput *iam.GetRoleOutput) {
	input := &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(pushApiWithInstance),
		ExecutionRoleArn: roleOutput.Role.Arn,
		NetworkMode: aws.String(ecs.NetworkModeAwsvpc),
		RequiresCompatibilities: []*string{aws.String(ecs.CompatibilityFargate)},
		Cpu: aws.String("256"),
		Memory: aws.String("512"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Cpu: aws.Int64(256),
				Image: aws.String(pushApiImage),
				MemoryReservation: aws.Int64(512),
				Name: aws.String(pushApi),
				//NetworkMode - TODO exists on terraform, but not here
				LogConfiguration: &ecs.LogConfiguration{
					LogDriver: aws.String(ecs.LogDriverAwslogs),
					Options: map[string]*string{
						"awslogs-group": aws.String(logsGroup),
						"awslogs-region": aws.String(awsRegion),
						"awslogs-stream-prefix": aws.String(logsStreamPrefix),
					},
				},
				PortMappings: []*ecs.PortMapping{
					{
						ContainerPort: aws.Int64(8080),
						HostPort: aws.Int64(8080),
					},
				},
				Environment: []*ecs.KeyValuePair{
					{
						Name: aws.String("PUSHAPI_REDIS__URL"),
						Value: aws.String("redis://" + pushRedisWithInstance + ".tsuru:6379"),
					},
					{
						Name: aws.String("PUSHAPI_PUSH_STREAM__URL"),
						Value: aws.String("http://" + pushStreamWithInstance + ".tsuru:9080"),
					},
				},
			},
		},
	}

	output, err := svc.RegisterTaskDefinition(input)
	if err != nil {
		fmt.Println("========== push-api - FAILED RegisterTaskDefinition ==========")
		panic(err)
	}
	fmt.Println("========== push-api - RegisterTaskDefinition ==========")
	fmt.Println(output.GoString())
}

func describePushApiService(svc *ecs.ECS) {
	input := &ecs.DescribeServicesInput{
		Cluster: aws.String(clusterName),
		Services: []*string{aws.String(pushApiWithInstance)},
	}

	output, err := svc.DescribeServices(input)
	if err != nil {
		fmt.Println("========== push-api - FAILED DescribeServices ==========")
		panic(err)
	}
	fmt.Println("========== push-api - DescribeServices ==========")
	fmt.Println(output.GoString())
}

func createPushApiService(svc *ecs.ECS, pushApiDiscovery *servicediscovery.CreateServiceOutput) {
	input := &ecs.CreateServiceInput{
		Cluster: aws.String(clusterName),
		DesiredCount: aws.Int64(1),
		ServiceName: aws.String(pushApiWithInstance),
		TaskDefinition: aws.String(pushApiWithInstance),
		LaunchType: aws.String(ecs.LaunchTypeFargate),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
				SecurityGroups: []*string{aws.String(sgServiceInboudAll), aws.String(sgServiceInboudOutboundSubnet)},
				Subnets: []*string{aws.String(subnet)},
			},
		},
		ServiceRegistries: []*ecs.ServiceRegistry{
			{
				RegistryArn: pushApiDiscovery.Service.Arn,
			},
		},
	}

	output, err := svc.CreateService(input)
	if err != nil {
		fmt.Println("========== push-api - FAILED CreateService ==========")
		panic(err)
	}
	fmt.Println("========== push-api - CreateService ==========")
	fmt.Println(output.GoString())
}

func deletePushApiService(svc *ecs.ECS) {
	input := &ecs.DeleteServiceInput{
		Cluster: aws.String(clusterName),
		Force: aws.Bool(true),
		Service: aws.String(pushApiWithInstance),
	}

	output, err := svc.DeleteService(input)
	if err != nil {
		fmt.Println("========== push-api - FAILED DeleteService ==========")
		panic(err)
	}
	fmt.Println("========== push-api - DeleteService ==========")
	fmt.Println(output.GoString())
}

func createPushApiServiceDiscovery(svc *servicediscovery.ServiceDiscovery) *servicediscovery.CreateServiceOutput {
	input := &servicediscovery.CreateServiceInput{
		Name: aws.String(pushApiWithInstance),
		NamespaceId: aws.String(dnsNamespace),
		DnsConfig: &servicediscovery.DnsConfig{
			DnsRecords: []*servicediscovery.DnsRecord{
				{
					TTL: aws.Int64(10),
					Type: aws.String("A"),
				},
			},
		},
		HealthCheckCustomConfig: &servicediscovery.HealthCheckCustomConfig{
			FailureThreshold: aws.Int64(1),
		},
	}

	output, err := svc.CreateService(input)
	if err != nil {
		fmt.Println("========== push-api - FAILED CreateService discovery ==========")
		panic(err)
	}
	fmt.Println("========== push-api - CreateService discovery ==========")
	fmt.Println(output.GoString())
	return output
}

func deletePushApiServiceDiscovery(svc *servicediscovery.ServiceDiscovery) {
	// TODO implement
}

///////////////////////////////////////////////////////////////////////////////
// push-stream
///////////////////////////////////////////////////////////////////////////////
func createPushStreamTaskDefinition(svc *ecs.ECS, roleOutput *iam.GetRoleOutput) {
	input := &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(pushStreamWithInstance),
		ExecutionRoleArn: roleOutput.Role.Arn,
		NetworkMode: aws.String(ecs.NetworkModeAwsvpc),
		RequiresCompatibilities: []*string{aws.String(ecs.CompatibilityFargate)},
		Cpu: aws.String("512"),
		Memory: aws.String("1024"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Cpu: aws.Int64(256),
				Image: aws.String(pushStreamImage),
				MemoryReservation: aws.Int64(512),
				Name: aws.String(pushStream),
				LogConfiguration: &ecs.LogConfiguration{
					LogDriver: aws.String(ecs.LogDriverAwslogs),
					Options: map[string]*string{
						"awslogs-group": aws.String(logsGroup),
						"awslogs-region": aws.String(awsRegion),
						"awslogs-stream-prefix": aws.String(logsStreamPrefix),
					},
				},
				PortMappings: []*ecs.PortMapping{
					{
						ContainerPort: aws.Int64(9080),
						HostPort: aws.Int64(9080),
					},
				},
			},
			{
				DependsOn: []*ecs.ContainerDependency{
					{
						Condition: aws.String(ecs.ContainerConditionSuccess),
						ContainerName: aws.String(pushStream),
					},
				},
				Cpu: aws.Int64(256),
				Image: aws.String(pushAgentImage),
				MemoryReservation: aws.Int64(512),
				Name: aws.String(pushAgent),
				LogConfiguration: &ecs.LogConfiguration{
					LogDriver: aws.String(ecs.LogDriverAwslogs),
					Options: map[string]*string{
						"awslogs-group": aws.String(logsGroup),
						"awslogs-region": aws.String(awsRegion),
						"awslogs-stream-prefix": aws.String(logsStreamPrefix),
					},
				},
				Environment: []*ecs.KeyValuePair{
					{
						Name: aws.String("PUSHAGENT_REDIS__URL"),
						Value: aws.String("redis://" + pushRedisWithInstance + ".tsuru:6379"),
					},
					{
						Name: aws.String("PUSHAGENT_PUSH_STREAM__URL"),
						Value: aws.String("http://" + pushStreamWithInstance + ".tsuru:9080"),
					},
				},
			},
		},
	}

	output, err := svc.RegisterTaskDefinition(input)
	if err != nil {
		fmt.Println("========== push-stream - FAILED RegisterTaskDefinition ==========")
		panic(err)
	}
	fmt.Println("========== push-stream - RegisterTaskDefinition ==========")
	fmt.Println(output.GoString())
}

func describePushStreamService(svc *ecs.ECS) {
	input := &ecs.DescribeServicesInput{
		Cluster: aws.String(clusterName),
		Services: []*string{aws.String(pushStreamWithInstance)},
	}

	output, err := svc.DescribeServices(input)
	if err != nil {
		fmt.Println("========== push-stream - FAILED DescribeServices ==========")
		panic(err)
	}
	fmt.Println("========== push-stream - DescribeServices ==========")
	fmt.Println(output.GoString())
}

func createPushStreamService(svc *ecs.ECS, pushStreamDiscovery *servicediscovery.CreateServiceOutput) {
	input := &ecs.CreateServiceInput{
		Cluster: aws.String(clusterName),
		DesiredCount: aws.Int64(1),
		ServiceName: aws.String(pushStreamWithInstance),
		TaskDefinition: aws.String(pushStreamWithInstance),
		LaunchType: aws.String(ecs.LaunchTypeFargate),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
				SecurityGroups: []*string{aws.String(sgServiceInboudAll), aws.String(sgServiceInboudOutboundSubnet)},
				Subnets: []*string{aws.String(subnet)},
			},
		},
		ServiceRegistries: []*ecs.ServiceRegistry{
			{
				RegistryArn: pushStreamDiscovery.Service.Arn,
			},
		},
	}

	output, err := svc.CreateService(input)
	if err != nil {
		fmt.Println("========== push-stream - FAILED CreateService ==========")
		panic(err)
	}
	fmt.Println("========== push-stream - CreateService ==========")
	fmt.Println(output.GoString())
}

func deletePushStreamService(svc *ecs.ECS) {
	input := &ecs.DeleteServiceInput{
		Cluster: aws.String(clusterName),
		Force: aws.Bool(true),
		Service: aws.String(pushStreamWithInstance),
	}

	output, err := svc.DeleteService(input)
	if err != nil {
		fmt.Println("========== push-stream - FAILED DeleteService ==========")
		panic(err)
	}
	fmt.Println("========== push-stream - DeleteService ==========")
	fmt.Println(output.GoString())
}

func createPushStreamServiceDiscovery(svc *servicediscovery.ServiceDiscovery) *servicediscovery.CreateServiceOutput {
	input := &servicediscovery.CreateServiceInput{
		Name: aws.String(pushStreamWithInstance),
		NamespaceId: aws.String(dnsNamespace),
		DnsConfig: &servicediscovery.DnsConfig{
			DnsRecords: []*servicediscovery.DnsRecord{
				{
					TTL: aws.Int64(10),
					Type: aws.String("A"),
				},
			},
		},
		HealthCheckCustomConfig: &servicediscovery.HealthCheckCustomConfig{
			FailureThreshold: aws.Int64(1),
		},
	}

	output, err := svc.CreateService(input)
	if err != nil {
		fmt.Println("========== push-stream - FAILED CreateService discovery ==========")
		panic(err)
	}
	fmt.Println("========== push-stream - CreateService discovery ==========")
	fmt.Println(output.GoString())
	return output
}

func deletePushStreamServiceDiscovery(svc *servicediscovery.ServiceDiscovery) {
	// TODO implement
}

///////////////////////////////////////////////////////////////////////////////
// service discovery
///////////////////////////////////////////////////////////////////////////////
func listServiceDiscoveryServices(svc *servicediscovery.ServiceDiscovery) *servicediscovery.ListServicesOutput {
	input := &servicediscovery.ListServicesInput{}

	output, err := svc.ListServices(input)
	if err != nil {
		fmt.Println("========== FAILED ListServices discovery ==========")
		panic(err)
	}
	fmt.Println("========== ListServices discovery ==========")
	fmt.Println(output.GoString())
	return output
}

func deleteServiceDiscoveryServices(svc *servicediscovery.ServiceDiscovery) {
	discoveryServices := listServiceDiscoveryServices(svc)
	services := discoveryServices.Services
	for _, s := range services {
		if strings.HasPrefix(*s.Name, "push-") {
			deleteInput := &servicediscovery.DeleteServiceInput{Id:s.Id}
			deleteOutput, err := svc.DeleteService(deleteInput)
			if err != nil {
				fmt.Println("========== FAILED DeleteService discovery ==========", *s.Name)
				panic(err)
			}
			fmt.Println("========== DeleteService discovery ==========")
			fmt.Println(deleteOutput.GoString())
		} else {
			fmt.Println("do not delete", s.Name)
		}
	}
}

func ignore(
	ecsSvc *ecs.ECS,
	discovery *servicediscovery.ServiceDiscovery,
	iamSvc *iam.IAM,
	output *iam.GetRoleOutput,
) {}

func main() {
	//pushaas.Run()
	const ACTION_LIST = "list"
	const ACTION_DESCRIBE = "describe"
	const ACTION_CREATE = "create"
	const ACTION_DELETE = "delete"
	action := os.Getenv("ACTION")

	mySession := session.Must(session.NewSession())
	ecsSvc := ecs.New(mySession)
	sdSvc := servicediscovery.New(mySession)
	iamSvc := iam.New(mySession)
	roleOutput := getIamRole(iamSvc)

	ignore(ecsSvc, sdSvc, iamSvc, roleOutput)

	if action == ACTION_LIST {
		//listTaskDescriptions(ecsSvc)
		//listServices(ecsSvc)
		listServiceDiscoveryServices(sdSvc)
		return
	}

	if action == ACTION_DESCRIBE {
		describeRedisService(ecsSvc)
		describePushApiService(ecsSvc)
		describePushStreamService(ecsSvc)
		return
	}

	if action == ACTION_CREATE {
		redisDiscovery := createRedisServiceDiscovery(sdSvc)
		createRedisService(ecsSvc, redisDiscovery)

		pushApiDiscovery := createPushApiServiceDiscovery(sdSvc)
		createPushApiTaskDefinition(ecsSvc, roleOutput)
		createPushApiService(ecsSvc, pushApiDiscovery)

		pushStreamDiscovery := createPushStreamServiceDiscovery(sdSvc)
		createPushStreamTaskDefinition(ecsSvc, roleOutput)
		createPushStreamService(ecsSvc, pushStreamDiscovery)
		return
	}

	if action == ACTION_DELETE {
		deleteRedisService(ecsSvc)
		deletePushApiService(ecsSvc)
		deletePushStreamService(ecsSvc)

		deleteServiceDiscoveryServices(sdSvc)
		return
	}
}
