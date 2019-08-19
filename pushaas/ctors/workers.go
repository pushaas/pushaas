package ctors

import (
	"github.com/RichardKnop/machinery/v1"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/provisioners"
	"github.com/rafaeleyng/pushaas/pushaas/services"
	"github.com/rafaeleyng/pushaas/pushaas/workers"
)

func NewProvisionWorker(config *viper.Viper, logger *zap.Logger, machineryServer *machinery.Server, provisioner provisioners.PushServiceProvisioner) workers.ProvisionWorker {
	return workers.NewProvisionWorker(config, logger, machineryServer, provisioner)
}

func NewInstanceWorker(config *viper.Viper, logger *zap.Logger, machineryServer *machinery.Server, instanceService services.InstanceService) workers.InstanceWorker {
	return workers.NewInstanceWorker(config, logger, machineryServer, instanceService)
}
