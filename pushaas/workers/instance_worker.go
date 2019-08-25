package workers

import (
	"encoding/json"
	"errors"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pushaas/pushaas/pushaas/models"
	"github.com/pushaas/pushaas/pushaas/provisioners"
	"github.com/pushaas/pushaas/pushaas/services"
)

type (
	InstanceWorker interface {
		HandleUpdateInstance(payload string) error
	}

	instanceWorker struct {
		logger                 *zap.Logger
		updateInstanceTaskName string
		instanceService        services.InstanceService
	}
)

func (w *instanceWorker) HandleUpdateInstance(payload string) error {
	var provisionResult provisioners.PushServiceProvisionResult
	err := json.Unmarshal([]byte(payload), &provisionResult)
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

func NewInstanceWorker(config *viper.Viper, logger *zap.Logger, instanceService services.InstanceService) InstanceWorker {
	return &instanceWorker{
		logger:                 logger.Named("instanceWorker"),
		updateInstanceTaskName: config.GetString("redis.pubsub.tasks.update_instance"),
		instanceService:        instanceService,
	}
}
