package provisioners

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type (
	awsEcsProvisioner struct {
		config *viper.Viper
		logger  *zap.Logger
	}
)

func (p awsEcsProvisioner) Provision(instanceName string) error {
	panic("implement me")
}

func (p awsEcsProvisioner) Deprovision(instanceName string) error {
	panic("implement me")
}

func NewAwsEcsProvisioner(config *viper.Viper, logger *zap.Logger) Provisioner {
	return &awsEcsProvisioner{
		config: config,
		logger: logger,
	}
}
