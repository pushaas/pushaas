package services

import (
	"encoding/json"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pushaas/pushaas/pushaas/models"
)

type (
	DispatchProvisionResult   int
	DispatchDeprovisionResult int

	ProvisionService interface {
		DispatchProvision(*models.Instance) DispatchProvisionResult
		DispatchDeprovision(*models.Instance) DispatchDeprovisionResult
	}

	provisionService struct {
		logger              *zap.Logger
		machineryServer     *machinery.Server
		provisionTaskName   string
		deprovisionTaskName string
	}
)

const (
	DispatchProvisionResultSuccess DispatchProvisionResult = iota
	DispatchProvisionResultFailure
)

const (
	DispatchDeprovisionResultSuccess DispatchDeprovisionResult = iota
	DispatchDeprovisionResultFailure
)

func (s *provisionService) buildProvisionSignature(messageJson *string) *tasks.Signature {
	return &tasks.Signature{
		Name: s.provisionTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: *messageJson,
			},
		},
	}
}

func (s *provisionService) DispatchProvision(instance *models.Instance) DispatchProvisionResult {
	bytes, err := json.Marshal(instance)
	if err != nil {
		s.logger.Error("error marshaling instance", zap.Any("instance", instance), zap.Error(err))
		return DispatchProvisionResultFailure
	}

	messageJson := string(bytes)
	signature := s.buildProvisionSignature(&messageJson)
	_, err = s.machineryServer.SendTask(signature)
	if err != nil {
		s.logger.Error("error dispatching provision for instance", zap.Any("instance", instance), zap.Error(err))
		return DispatchProvisionResultFailure
	}

	s.logger.Debug("instance provision dispatched", zap.Any("instance", instance))
	return DispatchProvisionResultSuccess
}

func (s *provisionService) buildDeprovisionSignature(messageJson string) *tasks.Signature {
	return &tasks.Signature{
		Name: s.deprovisionTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: messageJson,
			},
		},
	}
}

func (s *provisionService) DispatchDeprovision(instance *models.Instance) DispatchDeprovisionResult {
	bytes, err := json.Marshal(instance)
	if err != nil {
		s.logger.Error("error marshaling instance", zap.Any("instance", instance), zap.Error(err))
		return DispatchDeprovisionResultFailure
	}

	messageJson := string(bytes)
	signature := s.buildDeprovisionSignature(messageJson)
	_, err = s.machineryServer.SendTask(signature)
	if err != nil {
		s.logger.Error("error dispatching deprovision for instance", zap.Any("instance", instance), zap.Error(err))
		return DispatchDeprovisionResultFailure
	}

	s.logger.Debug("instance deprovision dispatched", zap.Any("instance", instance))
	return DispatchDeprovisionResultSuccess
}

func NewProvisionService(config *viper.Viper, logger *zap.Logger, machineryServer *machinery.Server) ProvisionService {
	return &provisionService{
		logger:              logger,
		machineryServer:     machineryServer,
		provisionTaskName:   config.GetString("redis.pubsub.tasks.provision"),
		deprovisionTaskName: config.GetString("redis.pubsub.tasks.deprovision"),
	}
}
