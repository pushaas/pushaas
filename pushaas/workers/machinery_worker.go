package workers

import (
	"github.com/RichardKnop/machinery/v1"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pushaas/pushaas/pushaas/services"
)

type (
	MachineryWorker interface {
		DispatchWorker()
	}

	machineryWorker struct {
		logger                 *zap.Logger
		machineryServer        *machinery.Server
		provisionTaskName      string
		deprovisionTaskName    string
		updateInstanceTaskName string
		instanceService        services.InstanceService
		enabled                bool
		provisionWorker        ProvisionWorker
		instanceWorker         InstanceWorker
	}
)

func (w *machineryWorker) startWorker() {
	w.logger.Info("starting worker")
	var err error

	err = w.machineryServer.RegisterTask(w.updateInstanceTaskName, w.instanceWorker.HandleUpdateInstance)
	if err != nil {
		w.logger.Error("failed to register update task", zap.Error(err))
		panic(err)
	}

	err = w.machineryServer.RegisterTask(w.provisionTaskName, w.provisionWorker.HandleProvisionTask)
	if err != nil {
		w.logger.Error("failed to register provision task", zap.Error(err))
		panic(err)
	}

	err = w.machineryServer.RegisterTask(w.deprovisionTaskName, w.provisionWorker.HandleDeprovisionTask)
	if err != nil {
		w.logger.Error("failed to register deprovision task", zap.Error(err))
		panic(err)
	}

	worker := w.machineryServer.NewWorker("worker", 0)
	err = worker.Launch()
	if err != nil {
		w.logger.Error("failed to launch worker", zap.Error(err))
		panic(err)
	}
}

func (w *machineryWorker) DispatchWorker() {
	if w.enabled {
		go w.startWorker()
		return
	}
	w.logger.Info("worker disabled, not starting")
}

func NewMachineryWorker(config *viper.Viper, logger *zap.Logger, machineryServer *machinery.Server, provisionWorker ProvisionWorker, instanceWorker InstanceWorker) MachineryWorker {
	enabled := config.GetBool("workers.machinery.enabled")
	workersEnabled := config.GetBool("workers.enabled")

	return &machineryWorker{
		logger:                 logger.Named("machineryWorker"),
		machineryServer:        machineryServer,
		provisionTaskName:      config.GetString("redis.pubsub.tasks.provision"),
		deprovisionTaskName:    config.GetString("redis.pubsub.tasks.deprovision"),
		updateInstanceTaskName: config.GetString("redis.pubsub.tasks.update-instance"),
		enabled:                enabled && workersEnabled,
		provisionWorker:        provisionWorker,
		instanceWorker:         instanceWorker,
	}
}
