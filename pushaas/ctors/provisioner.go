package ctors

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/provisioners"
	"github.com/rafaeleyng/pushaas/pushaas/provisioners/aws_ecs_provisioner"
)

func NewProvisioner(config *viper.Viper, logger *zap.Logger) (provisioners.Provisioner, error) {
	provider := config.GetString("provisioner.provider")

	if provider == "aws-ecs" {
		logger.Info("initializing provisioner with provider", zap.String("provider", provider))
		return aws_ecs_provisioner.NewAwsEcsProvisioner(config, logger)
	}

	return nil, fmt.Errorf("unknown provider: %s", provider)
}
