package ctors

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/provisioners"
	"github.com/rafaeleyng/pushaas/pushaas/provisioners/ecs_provisioner"
)

func NewPushServiceProvisioner(
	config *viper.Viper,
	logger *zap.Logger,
	pushRedisProvisioner ecs_provisioner.EcsPushRedisProvisioner,
	pushStreamProvisioner ecs_provisioner.EcsPushStreamProvisioner,
	pushApiProvisioner ecs_provisioner.EcsPushApiProvisioner,
) (provisioners.PushServiceProvisioner, error) {
	provider := config.GetString("provisioner.provider")

	if provider == "ecs" {
		logger.Info("initializing provisioner with provider", zap.String("provider", provider))
		return ecs_provisioner.NewEcsPushServiceProvisioner(config, logger, pushRedisProvisioner, pushStreamProvisioner, pushApiProvisioner)
	}

	return nil, fmt.Errorf("unknown provider: %s", provider)
}

/*
	ecs
*/
func NewEcsPushRedisProvisioner() ecs_provisioner.EcsPushRedisProvisioner {
	return ecs_provisioner.NewEcsPushRedisProvisioner()
}

func NewEcsPushStreamProvisioner() ecs_provisioner.EcsPushStreamProvisioner {
	return ecs_provisioner.NewEcsPushStreamProvisioner()
}

func NewEcsPushApiProvisioner() ecs_provisioner.EcsPushApiProvisioner {
	return ecs_provisioner.NewEcsPushApiProvisioner()
}
