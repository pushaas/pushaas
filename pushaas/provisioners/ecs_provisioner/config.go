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
	imagePushApi := config.GetString("provisioner.ecs.image-push-api")
	imagePushAgent := config.GetString("provisioner.ecs.image-push-agent")
	imagePushStream := config.GetString("provisioner.ecs.image-push-stream")

	region := config.GetString("provisioner.ecs.region")
	cluster := config.GetString("provisioner.ecs.cluster")
	logsStreamPrefix := config.GetString("provisioner.ecs.logs-stream-prefix")
	logsGroup := config.GetString("provisioner.ecs.logs-group")

	// required vars
	// comes from `https://github.com/pushaas/pushaas-aws-ecs-config/scripts/40-pushaas/30-create-cluster/terraform.tfstate`
	securityGroup := config.GetString("provisioner.ecs.security-group")
	// comes from `https://github.com/pushaas/pushaas-aws-ecs-config/scripts/10-vpc/10-create-vpc/terraform.tfstate`
	subnet := config.GetString("provisioner.ecs.subnet")
	// comes from `https://github.com/pushaas/pushaas-aws-ecs-config/scripts/30-dns/10-create-namespace/terraform.tfstate`
	dnsNamespace := config.GetString("provisioner.ecs.dns-namespace")

	requiredVars := map[string]string{
		"provisioner.ecs.security-group": securityGroup,
		"provisioner.ecs.subnet": subnet,
		"provisioner.ecs.dns-namespace": dnsNamespace,
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
