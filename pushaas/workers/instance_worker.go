package workers

import (
	"encoding/json"
	"errors"

	"github.com/RichardKnop/machinery/v1"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/models"
	"github.com/rafaeleyng/pushaas/pushaas/provisioners"
	"github.com/rafaeleyng/pushaas/pushaas/services"
)

type (
	InstanceWorker interface {
		DispatchWorker()
	}

	instanceWorker struct {
		logger                 *zap.Logger
		machineryServer        *machinery.Server
		updateInstanceTaskName string
		instanceService        services.InstanceService
		enabled                bool
	}
)

func (w *instanceWorker) handleUpdateInstance(payload string) error {
	var provisionResult *provisioners.PushServiceProvisionResult
	err := json.Unmarshal([]byte(payload), provisionResult)
	if err != nil {
		w.logger.Error("failed to unmarshal instance to update", zap.String("payload", payload), zap.Error(err))
		return err
	}

	instanceName := provisionResult.Instance.Name

	// if failed to provision
	if provisionResult.Status == provisioners.PushServiceProvisionStatusFailure {
		updateResult := w.instanceService.UpdateStatus(instanceName, models.InstanceStatusFailed)
		if updateResult == services.InstanceUpdateFailure {
			w.logger.Error("failed to update instance status after failure", zap.Any("provisionResult", provisionResult))
			return errors.New("failed to update instance status after failure")
		}
		return nil
	}

	// if succeeded to provision
	updateResult := w.instanceService.UpdateStatus(instanceName, models.InstanceStatusRunning)
	if updateResult == services.InstanceUpdateFailure {
		w.logger.Error("failed to update instance status after success", zap.Any("provisionResult", provisionResult))
		return errors.New("failed to update instance status after success")
	}

	_, err = w.instanceService.SetInstanceVars(instanceName, provisionResult.EnvVars)
	if err != nil {
		w.logger.Error("failed to set instance variables after success", zap.Any("provisionResult", provisionResult), zap.Error(err))
		return errors.New("failed to set instance variables after success")
	}

	return nil
}

func (w *instanceWorker) startWorker() {
	w.logger.Info("starting worker")
	var err error

	err = w.machineryServer.RegisterTask(w.updateInstanceTaskName, w.handleUpdateInstance)
	if err != nil {
		w.logger.Error("failed to register update task", zap.Error(err))
		panic(err)
	}

	worker := w.machineryServer.NewWorker("instance_worker", 0)
	err = worker.Launch()
	if err != nil {
		w.logger.Error("failed to launch instance_worker", zap.Error(err))
		panic(err)
	}
}

func (w *instanceWorker) DispatchWorker() {
	if w.enabled {
		go w.startWorker()
	} else {
		w.logger.Info("worker disabled, not starting")
	}
}

func NewInstanceWorker(config *viper.Viper, logger *zap.Logger, machineryServer *machinery.Server, instanceService services.InstanceService) InstanceWorker {
	enabled := config.GetBool("workers.instance.enabled")
	workersEnabled := config.GetBool("workers.enabled")

	return &instanceWorker{
		logger:                 logger.Named("instanceWorker"),
		machineryServer:        machineryServer,
		updateInstanceTaskName: config.GetString("redis.pubsub.tasks.update-instance"),
		enabled:                enabled && workersEnabled,
		instanceService:        instanceService,
	}
}
