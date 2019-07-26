package services

import (
	"fmt"

	"github.com/fatih/structs"
	"github.com/go-redis/redis"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/models"
)

type (
	InstanceCreationResult  int
	InstanceRetrievalResult int
	InstanceDeletionResult  int
	InstanceStatusResult    int

	InstanceService interface {
		Create(instanceForm *models.InstanceForm) InstanceCreationResult
		GetByName(name string) (*models.Instance, InstanceRetrievalResult)
		Delete(name string) InstanceDeletionResult
		GetStatusByName(name string) InstanceStatusResult
	}

	instanceService struct {
		instanceKeyPrefix  string
		bindingsKeyPrefix  string
		unitsHostKeyPrefix string

		logger           *zap.Logger
		provisionService ProvisionService
		redisClient      redis.UniversalClient
	}
)

const (
	InstanceRetrievalSuccess InstanceRetrievalResult = iota
	InstanceRetrievalNotFound
	InstanceRetrievalFailure
)

const (
	InstanceCreationSuccess InstanceCreationResult = iota
	InstanceCreationAlreadyExist
	InstanceCreationInvalidData
	InstanceCreationFailure
	InstanceCreationProvisionFailure
)

const (
	InstanceDeletionSuccess InstanceDeletionResult = iota
	InstanceDeletionNotFound
	InstanceDeletionFailure
	InstanceDeletionDeprovisionFailure
)

const (
	InstanceStatusNotFound InstanceStatusResult = iota
	InstanceStatusFailure

	InstanceStatusRunningStatus
	InstanceStatusPendingStatus
	InstanceStatusFailedStatus
)

func prefixKey(prefix, value string) string {
	return fmt.Sprintf("%s%s", prefix, value)
}

func (s *instanceService) instanceKey(instanceName string) string {
	return prefixKey(s.instanceKeyPrefix, instanceName)
}

func (s *instanceService) bindingsKey(instanceName string) string {
	return prefixKey(s.bindingsKeyPrefix, instanceName)
}

func (s *instanceService) unitsHostKey(instanceName, appHost, appName string) string {
	return prefixKey(s.unitsHostKeyPrefix, instanceName)
}

func (s *instanceService) GetByName(instanceName string) (*models.Instance, InstanceRetrievalResult) {
	var err error
	instanceKey := s.instanceKey(instanceName)

	// retrieve
	cmd := s.redisClient.HGetAll(instanceKey)
	instanceMap, err := cmd.Result()
	if err != nil {
		s.logger.Error("failed to retrieve instance", zap.Error(err), zap.String("instance", instanceName))
		return nil, InstanceRetrievalFailure
	}
	if len(instanceMap) == 0 {
		s.logger.Info("instance not found", zap.String("instance", instanceName))
		return nil, InstanceRetrievalNotFound
	}

	// decode
	var instance models.Instance
	err = mapstructure.Decode(instanceMap, &instance)
	if err != nil {
		s.logger.Error("failed to decode instance", zap.Error(err), zap.String("instance", instanceName))
		return nil, InstanceRetrievalFailure
	}

	return &instance, InstanceRetrievalSuccess
}

func (s *instanceService) doCreate(instance *models.Instance) InstanceCreationResult {
	instanceName := instance.Name
	instance.Status = models.InstanceStatusPending

	// create
	instanceKey := s.instanceKey(instanceName)
	// TODO decide how to add Bindings and UnitsHosts
	//bindingsKey := s.bindingsKey(instanceName)
	//unitsHostKey := s.unitsHostKey(instanceName)

	//pipeline := s.redisClient.TxPipeline()
	//pipeline.HMSet(instanceKey, instanceMap)
	////pipeline.
	//cmders, err := pipeline.Exec()

	var err error
	instanceMap := structs.Map(instance)
	err = s.redisClient.HMSet(instanceKey, instanceMap).Err()
	if err != nil {
		s.logger.Error("failed to create instance", zap.Error(err), zap.Any("instance", instance))
		return InstanceCreationFailure
	}
	return InstanceCreationSuccess
}

func (s *instanceService) Create(instanceForm *models.InstanceForm) InstanceCreationResult {
	instanceName := instanceForm.Name

	// check existing
	_, resultGet := s.GetByName(instanceName)
	if resultGet == InstanceRetrievalSuccess {
		return InstanceCreationAlreadyExist
	} else if resultGet == InstanceRetrievalFailure {
		return InstanceCreationFailure
	}

	// validate
	validationResult := instanceForm.Validate()
	if validationResult == models.InstanceFormInvalid {
		return InstanceCreationInvalidData
	}
	instance := models.InstanceFromInstanceForm(instanceForm)

	// create
	resultCreate := s.doCreate(instance)
	if resultCreate != InstanceCreationSuccess {
		return resultCreate
	}

	// dispatch provision
	dispatchProvisionResult := s.provisionService.DispatchProvision(instance)
	if dispatchProvisionResult != DispatchProvisionResultSuccess {
		s.logger.Error("failed to dispatch provision", zap.Any("instance", instance))
		return InstanceCreationProvisionFailure
	}

	return InstanceCreationSuccess
}

func (s *instanceService) doDelete(instance *models.Instance) InstanceDeletionResult {
	instanceName := instance.Name
	instanceKey := s.instanceKey(instanceName)

	// TODO delete Bindings and UnitsHosts as set or list or other thing
	var err error
	value, err := s.redisClient.Del(instanceKey).Result()

	if err != nil {
		s.logger.Error("error while trying to delete instance", zap.String("name", instanceName), zap.Error(err))
		return InstanceDeletionFailure
	} else if value == 0 {
		s.logger.Error("instance not found to be deleted", zap.String("name", instanceName))
		return InstanceDeletionNotFound
	}
	return InstanceDeletionSuccess
}

func (s *instanceService) Delete(instanceName string) InstanceDeletionResult {
	// check existing
	instance, resultGet := s.GetByName(instanceName)
	if resultGet == InstanceRetrievalNotFound {
		return InstanceDeletionNotFound
	} else if resultGet == InstanceRetrievalFailure {
		return InstanceDeletionFailure
	}

	// delete
	resultDelete := s.doDelete(instance)
	if resultDelete != InstanceDeletionSuccess {
		return resultDelete
	}

	// deprovision
	dispatchDeprovisionResult := s.provisionService.DispatchDeprovision(instance)
	if dispatchDeprovisionResult != DispatchDeprovisionResultSuccess {
		s.logger.Error("failed to dispatch deprovision", zap.Any("instance", instance))
		return InstanceDeletionDeprovisionFailure
	}

	return InstanceDeletionSuccess
}

func (s *instanceService) GetStatusByName(name string) InstanceStatusResult {
	// retrieve
	instance, resultGet := s.GetByName(name)
	if resultGet == InstanceRetrievalNotFound {
		s.logger.Error("instance not found to check status", zap.String("name", name))
		return InstanceStatusNotFound
	} else if resultGet == InstanceRetrievalFailure {
		s.logger.Error("failed to get instance to check status", zap.String("name", name))
		return InstanceStatusFailure
	}

	// check status
	if instance.Status == models.InstanceStatusPending {
		return InstanceStatusPendingStatus
	} else if instance.Status == models.InstanceStatusFailed {
		return InstanceStatusFailedStatus
	}
	return InstanceStatusRunningStatus
}

func NewInstanceService(config *viper.Viper, logger *zap.Logger, redisClient redis.UniversalClient, provisionService ProvisionService) InstanceService {
	instanceKeyPrefix := config.GetString("redis.db.instance.prefix")
	bindingsKeyPrefix := config.GetString("redis.db.bindings.prefix")
	unitsHostKeyPrefix := config.GetString("redis.db.units-host.prefix")

	return &instanceService{
		instanceKeyPrefix:  instanceKeyPrefix,
		bindingsKeyPrefix:  bindingsKeyPrefix,
		unitsHostKeyPrefix: unitsHostKeyPrefix,

		logger:           logger,
		provisionService: provisionService,
		redisClient:      redisClient,
	}
}
