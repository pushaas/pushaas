package workers

import (
	"encoding/json"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pushaas/pushaas/pushaas/models"
	"github.com/pushaas/pushaas/pushaas/provisioners"
)

type (
	ProvisionWorker interface {
		HandleProvisionTask(payload string) error
		HandleDeprovisionTask(payload string) error
	}

	provisionWorker struct {
		logger                 *zap.Logger
		machineryServer        *machinery.Server
		updateInstanceTaskName string
		provisioner            provisioners.PushServiceProvisioner
	}
)

func (w *provisionWorker) buildUpdateInstanceSignature(messageJson string) *tasks.Signature {
	return &tasks.Signature{
		Name: w.updateInstanceTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: messageJson,
			},
		},
	}
}

func (w *provisionWorker) sendUpdateTask(provisionResult *provisioners.PushServiceProvisionResult) error {
	bytes, err := json.Marshal(provisionResult)
	if err != nil {
		w.logger.Error("error marshaling provisionResult", zap.Any("provisionResult", provisionResult), zap.Error(err))
		return err
	}

	messageJson := string(bytes)
	signature := w.buildUpdateInstanceSignature(messageJson)
	_, err = w.machineryServer.SendTask(signature)
	if err != nil {
		w.logger.Error("error dispatching update for instance", zap.Any("provisionResult", provisionResult), zap.Error(err))
		return err
	}

	w.logger.Debug("instance update dispatched", zap.Any("provisionResult", provisionResult))
	return nil
}

func (w *provisionWorker) HandleProvisionTask(payload string) error {
	var instance PushServiceProvisionResult
	err := json.Unmarshal([]byte(payload), &instance)
	if err != nil {
		w.logger.Error("failed to unmarshal instance to provision", zap.String("payload", payload), zap.Error(err))
		return err
	}

	provisionResult := w.provisioner.Provision(&instance)
	return w.sendUpdateTask(provisionResult)
}

func (w *provisionWorker) HandleDeprovisionTask(payload string) error {
	var instance models.Instance
	err := json.Unmarshal([]byte(payload), &instance)
	if err != nil {
		w.logger.Error("failed to unmarshal instance to deprovision", zap.String("payload", payload), zap.Error(err))
		return err
	}

	w.provisioner.Deprovision(&instance)
	return nil
}

func NewProvisionWorker(config *viper.Viper, logger *zap.Logger, machineryServer *machinery.Server, provisioner provisioners.PushServiceProvisioner) ProvisionWorker {
	return &provisionWorker{
		logger:                 logger.Named("provisionWorker"),
		machineryServer:        machineryServer,
		updateInstanceTaskName: config.GetString("redis.pubsub.tasks.update_instance"),
		provisioner:            provisioner,
	}
}
