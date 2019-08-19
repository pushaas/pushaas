package workers

import (
	"encoding/json"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/models"
	"github.com/rafaeleyng/pushaas/pushaas/provisioners"
)

type (
	ProvisionWorker interface {
		DispatchWorker()
	}

	provisionWorker struct {
		logger                 *zap.Logger
		machineryServer        *machinery.Server
		provisionTaskName      string
		deprovisionTaskName    string
		updateInstanceTaskName string
		provisioner            provisioners.PushServiceProvisioner
		enabled                bool
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

func (w *provisionWorker) handleProvisionTask(payload string) error {
	var instance models.Instance
	err := json.Unmarshal([]byte(payload), &instance)
	if err != nil {
		w.logger.Error("failed to unmarshal instance to provision", zap.String("payload", payload), zap.Error(err))
		return err
	}

	provisionResult := w.provisioner.Provision(&instance)
	return w.sendUpdateTask(provisionResult)
}

func (w *provisionWorker) handleDeprovisionTask(payload string) error {
	var instance models.Instance
	err := json.Unmarshal([]byte(payload), &instance)
	if err != nil {
		w.logger.Error("failed to unmarshal instance to deprovision", zap.String("payload", payload), zap.Error(err))
		return err
	}

	w.provisioner.Deprovision(&instance)
	return nil
}

func (w *provisionWorker) startWorker() {
	w.logger.Info("starting worker")
	var err error

	err = w.machineryServer.RegisterTask(w.provisionTaskName, w.handleProvisionTask)
	if err != nil {
		w.logger.Error("failed to register provision task", zap.Error(err))
		panic(err)
	}

	err = w.machineryServer.RegisterTask(w.deprovisionTaskName, w.handleDeprovisionTask)
	if err != nil {
		w.logger.Error("failed to register deprovision task", zap.Error(err))
		panic(err)
	}

	worker := w.machineryServer.NewWorker("provision_worker", 0)
	err = worker.Launch()
	if err != nil {
		w.logger.Error("failed to launch provision_worker", zap.Error(err))
		panic(err)
	}
}

func (w *provisionWorker) DispatchWorker() {
	if w.enabled {
		go w.startWorker()
	} else {
		w.logger.Info("worker disabled, not starting")
	}
}

func NewProvisionWorker(config *viper.Viper, logger *zap.Logger, machineryServer *machinery.Server, provisioner provisioners.PushServiceProvisioner) ProvisionWorker {
	enabled := config.GetBool("workers.provision.enabled")
	workersEnabled := config.GetBool("workers.enabled")

	return &provisionWorker{
		logger:                 logger.Named("provisionWorker"),
		machineryServer:        machineryServer,
		provisionTaskName:      config.GetString("redis.pubsub.tasks.provision"),
		deprovisionTaskName:    config.GetString("redis.pubsub.tasks.deprovision"),
		updateInstanceTaskName: config.GetString("redis.pubsub.tasks.update-instance"),
		enabled:                enabled && workersEnabled,
		provisioner:            provisioner,
	}
}
