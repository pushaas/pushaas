package provisioners

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/models"
)

//import "github.com/aws/aws-sdk-go/service/ecs"

type (
	awsEcsProvisioner struct {
		config *viper.Viper
		logger  *zap.Logger
	}
)

func (p awsEcsProvisioner) Provision(instance *models.Instance) ProvisionResult {
	p.logger.Info("######## did call Provision")
	return ProvisionResultSuccess
}

func (p awsEcsProvisioner) Deprovision(instance *models.Instance) DeprovisionResult {
	p.logger.Info("######## did call Deprovision")
	return DeprovisionResultSuccess
}

func NewAwsEcsProvisioner(config *viper.Viper, logger *zap.Logger) Provisioner {
	return &awsEcsProvisioner{
		config: config,
		logger: logger,
	}
}
