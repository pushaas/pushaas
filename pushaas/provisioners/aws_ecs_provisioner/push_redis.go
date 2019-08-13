package aws_ecs_provisioner

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/servicediscovery"

	"github.com/rafaeleyng/pushaas/pushaas/models"
)

const pushRedis = "push-redis"

func pushRedisWithInstance(instanceName string) string {
	return pushRedis + "-" + instanceName
}

func createRedisServiceDiscovery(
	instance *models.Instance,
	serviceDiscoverySvc *servicediscovery.ServiceDiscovery,
	provisionerConfig *awsEcsProvisionerConfig,
) (*servicediscovery.CreateServiceOutput, error) {
	return serviceDiscoverySvc.CreateService(&servicediscovery.CreateServiceInput{
		Name:        aws.String(pushRedisWithInstance(instance.Name)),
		NamespaceId: aws.String(provisionerConfig.dnsNamespace),
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

func createRedisService(
	instance *models.Instance,
	ecsSvc *ecs.ECS,
	redisDiscovery *servicediscovery.CreateServiceOutput,
	provisionerConfig *awsEcsProvisionerConfig,
) (*ecs.CreateServiceOutput, error) {
	return ecsSvc.CreateService(&ecs.CreateServiceInput{
		Cluster:        aws.String(provisionerConfig.cluster),
		DesiredCount:   aws.Int64(1),
		ServiceName:    aws.String(pushRedisWithInstance(instance.Name)),
		TaskDefinition: aws.String(pushRedis),
		LaunchType:     aws.String(ecs.LaunchTypeFargate),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
				SecurityGroups: []*string{aws.String(provisionerConfig.securityGroup)},
				Subnets:        []*string{aws.String(provisionerConfig.subnet)},
			},
		},
		ServiceRegistries: []*ecs.ServiceRegistry{
			{
				RegistryArn: redisDiscovery.Service.Arn,
			},
		},
	})
}

//func deleteRedisService(svc *ecs.ECS) {
//	input := &ecs.DeleteServiceInput{
//		Cluster: aws.String(clusterName),
//		Force:   aws.Bool(true),
//		Service: aws.String(pushRedisWithInstance),
//	}
//
//	output, err := svc.DeleteService(input)
//	if err != nil {
//		fmt.Println("========== redis - FAILED DeleteService ==========")
//		panic(err)
//	}
//	fmt.Println("========== redis - DeleteService ==========")
//	fmt.Println(output.GoString())
//}
