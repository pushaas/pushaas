package ecs_provisioner

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/servicediscovery/servicediscoveryiface"
	"github.com/spf13/viper"
)

type (
	EcsProvisionerConfig struct {
		ecs              ecsiface.ECSAPI
		iam              iamiface.IAMAPI
		ec2              ec2iface.EC2API
		serviceDiscovery servicediscoveryiface.ServiceDiscoveryAPI
		imagePushApi     *string
		imagePushAgent   *string
		imagePushStream  *string
		region           *string
		cluster          *string
		logsStreamPrefix *string
		logsGroup        *string
		securityGroup    *string
		subnet           *string
		dnsNamespace     *string
	}
)

func NewEcsProvisionerConfig(config *viper.Viper, iamSvc iamiface.IAMAPI, ecsSvc ecsiface.ECSAPI, ec2Svc ec2iface.EC2API, serviceDiscoverySvc servicediscoveryiface.ServiceDiscoveryAPI) (*EcsProvisionerConfig, error) {
	imagePushApi := config.GetString("provisioner.ecs.image_push_api")
	imagePushAgent := config.GetString("provisioner.ecs.image_push_agent")
	imagePushStream := config.GetString("provisioner.ecs.image_push_stream")

	region := config.GetString("provisioner.ecs.region")
	cluster := config.GetString("provisioner.ecs.cluster")
	logsStreamPrefix := config.GetString("provisioner.ecs.logs_stream_prefix")
	logsGroup := config.GetString("provisioner.ecs.logs_group")

	// value configured in `https://github.com/pushaas/pushaas-aws-ecs-config/scripts/40-pushaas/30-create-cluster/terraform.tfstate`
	securityGroupKey := "provisioner.ecs.security_group"
	securityGroup := config.GetString(securityGroupKey)

	// value configured in `https://github.com/pushaas/pushaas-aws-ecs-config/scripts/10-vpc/10-create-vpc/terraform.tfstate`
	subnetKey := "provisioner.ecs.subnet"
	subnet := config.GetString(subnetKey)

	// value configured in `https://github.com/pushaas/pushaas-aws-ecs-config/scripts/30-dns/10-create-namespace/terraform.tfstate`
	dnsNamespaceKey := "provisioner.ecs.dns_namespace"
	dnsNamespace := config.GetString(dnsNamespaceKey)

	requiredVars := map[string]string{
		securityGroupKey: securityGroup,
		subnetKey: subnet,
		dnsNamespaceKey: dnsNamespace,
	}

	for k, v := range requiredVars {
		if v == "" {
			return nil, errors.New(fmt.Sprintf("ecsProvisioner config required and not set: %s", k))
		}
	}

	return &EcsProvisionerConfig{
		iam:              iamSvc,
		ecs:              ecsSvc,
		ec2:              ec2Svc,
		serviceDiscovery: serviceDiscoverySvc,
		imagePushApi:     aws.String(imagePushApi),
		imagePushAgent:   aws.String(imagePushAgent),
		imagePushStream:  aws.String(imagePushStream),
		region:           aws.String(region),
		cluster:          aws.String(cluster),
		logsStreamPrefix: aws.String(logsStreamPrefix),
		logsGroup:        aws.String(logsGroup),
		securityGroup:    aws.String(securityGroup),
		subnet:           aws.String(subnet),
		dnsNamespace:     aws.String(dnsNamespace),
	}, nil
}
