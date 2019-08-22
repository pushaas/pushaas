package ctors

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pushaas/pushaas/pushaas/provisioners"
	"github.com/pushaas/pushaas/pushaas/provisioners/ecs_provisioner"
)

func NewPushServiceProvisioner(
	config *viper.Viper,
	logger *zap.Logger,
	provisionerConfig *ecs_provisioner.EcsProvisionerConfig,
	pushRedisProvisioner ecs_provisioner.EcsPushRedisProvisioner,
	pushStreamProvisioner ecs_provisioner.EcsPushStreamProvisioner,
	pushApiProvisioner ecs_provisioner.EcsPushApiProvisioner,
) (provisioners.PushServiceProvisioner, error) {
	provider := config.GetString("provisioner.provider")

	if provider == "ecs" {
		logger.Info("initializing provisioner with provider", zap.String("provider", provider))
		return ecs_provisioner.NewEcsPushServiceProvisioner(logger, provisionerConfig, pushRedisProvisioner, pushStreamProvisioner, pushApiProvisioner)
	}

	return nil, fmt.Errorf("unknown provider: %s", provider)
}

/*
	aws ecs
*/
func NewEcsProvisionerConfig(config *viper.Viper) (*ecs_provisioner.EcsProvisionerConfig, error) {
	awsSession := session.Must(session.NewSession())
	iamSvc := iam.New(awsSession)
	ecsSvc := ecs.New(awsSession)
	ec2Svc := ec2.New(awsSession)
	serviceDiscoverySvc := servicediscovery.New(awsSession)
	return ecs_provisioner.NewEcsProvisionerConfig(config, iamSvc, ecsSvc, ec2Svc, serviceDiscoverySvc)
}

func NewEcsPushRedisProvisioner(logger *zap.Logger, ecsConfig *ecs_provisioner.EcsProvisionerConfig) ecs_provisioner.EcsPushRedisProvisioner {
	return ecs_provisioner.NewEcsPushRedisProvisioner(logger, ecsConfig)
}

func NewEcsPushStreamProvisioner(logger *zap.Logger, ecsConfig *ecs_provisioner.EcsProvisionerConfig) ecs_provisioner.EcsPushStreamProvisioner {
	return ecs_provisioner.NewEcsPushStreamProvisioner(logger, ecsConfig)
}

func NewEcsPushApiProvisioner(logger *zap.Logger, ecsConfig *ecs_provisioner.EcsProvisionerConfig) ecs_provisioner.EcsPushApiProvisioner {
	return ecs_provisioner.NewEcsPushApiProvisioner(logger, ecsConfig)
}
