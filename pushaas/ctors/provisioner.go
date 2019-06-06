package ctors

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/provisioners"
)

func NewProvisioner(config *viper.Viper, logger *zap.Logger) (provisioners.Provisioner, error) {
	provider := config.GetString("provisioner.provider")

	if provider == "aws-ecs" {
		logger.Info("initializing provisioner with provider", zap.String("provider", provider))
		return provisioners.NewAwsEcsProvisioner(config, logger), nil
	}

	return nil, fmt.Errorf("unknown provider: %s", provider)
}
