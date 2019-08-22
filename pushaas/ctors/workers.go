package ctors

import (
	"github.com/RichardKnop/machinery/v1"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pushaas/pushaas/pushaas/provisioners"
	"github.com/pushaas/pushaas/pushaas/services"
	"github.com/pushaas/pushaas/pushaas/workers"
)

func NewProvisionWorker(config *viper.Viper, logger *zap.Logger, machineryServer *machinery.Server, provisioner provisioners.PushServiceProvisioner) workers.ProvisionWorker {
	return workers.NewProvisionWorker(config, logger, machineryServer, provisioner)
}

func NewInstanceWorker(config *viper.Viper, logger *zap.Logger, instanceService services.InstanceService) workers.InstanceWorker {
	return workers.NewInstanceWorker(config, logger, instanceService)
}

func NewMachineryWorker(config *viper.Viper, logger *zap.Logger, machineryServer *machinery.Server, provisionWorker workers.ProvisionWorker, instanceWorker workers.InstanceWorker) workers.MachineryWorker {
	return workers.NewMachineryWorker(config, logger, machineryServer, provisionWorker, instanceWorker)
}

