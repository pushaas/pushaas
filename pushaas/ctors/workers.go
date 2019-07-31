package ctors

import (
	"github.com/spf13/viper"

	"github.com/rafaeleyng/pushaas/pushaas/workers"
)

func NewProvisionWorker(config *viper.Viper) workers.ProvisionWorker {
	return workers.NewProvisionWorker(config)
}
