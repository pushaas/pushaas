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
	InstanceUpdateResult    int

	InstanceService interface {
		Create(instanceForm *models.InstanceForm) InstanceCreationResult
		GetAll() ([]*models.Instance, InstanceRetrievalResult)
		GetByName(name string) (*models.Instance, InstanceRetrievalResult)
		Delete(name string) InstanceDeletionResult
		UpdateStatus(name string, status models.InstanceStatus) InstanceUpdateResult
		GetStatusByName(name string) InstanceStatusResult
		GetInstanceVars(name string) (map[string]string, error)
		SetInstanceVars(name string, envVars map[string]string) (string, error)
		DelInstanceVars(name string) (int64, error)
	}

	instanceService struct {
		instanceKeyPrefix     string
		instanceVarsKeyPrefix string
		logger                *zap.Logger
		provisionService      ProvisionService
		redisClient           redis.UniversalClient
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
	InstanceUpdateSuccess InstanceUpdateResult = iota
	InstanceUpdateFailure
)

const (
	InstanceStatusNotFound InstanceStatusResult = iota
	InstanceStatusFailure

	InstanceStatusRunningStatus
	InstanceStatusPendingStatus
	InstanceStatusFailedStatus
)

/*
	===========================================================================
	instances
	===========================================================================
*/
func (s *instanceService) instanceKey(instanceName string) string {
	return fmt.Sprintf("%s:%s", s.instanceKeyPrefix, instanceName)
}

func (s *instanceService) GetAll() ([]*models.Instance, InstanceRetrievalResult) {
	var err error
	patternAllInstanceKeys := s.instanceKey("*")

	// get keys
	keys, err := s.redisClient.Keys(patternAllInstanceKeys).Result()
	if err != nil {
		s.logger.Error("failed to retrieve instance keys", zap.Error(err), zap.String("patternAllInstanceKeys", patternAllInstanceKeys))
		return nil, InstanceRetrievalFailure
	}
	if len(keys) == 0 {
		return nil, InstanceRetrievalNotFound
	}

	//s.logger.Info("TODO", zap.Any("keys", keys))
	//return nil, InstanceRetrievalSuccess

	// init pipeline
	pipeline := s.redisClient.Pipeline()
	defer func() {
		err := pipeline.Close()
		if err != nil {
			s.logger.Error("failed to close pipeline", zap.Error(err))
		}
	}()

	// fill pipeline commands
	for _, key := range keys {
		pipeline.HGetAll(key)
	}

	// exec pipeline
	results, err := pipeline.Exec()
	if err != nil {
		s.logger.Error("failed to execute pipeline to retrieve instances", zap.Error(err))
		return nil, InstanceRetrievalFailure
	}

	instances := make([]*models.Instance, len(keys))
	for i, cmd := range results {
		mapCmd := cmd.(*redis.StringStringMapCmd)
		instanceMap, err := mapCmd.Result()
		if err != nil {
			s.logger.Error("failed to retrieve instance", zap.Error(err), zap.String("key", keys[i]))
			return nil, InstanceRetrievalFailure
		}
		if len(instanceMap) == 0 {
			return nil, InstanceRetrievalNotFound
		}

		// decode
		var instance models.Instance
		err = mapstructure.Decode(instanceMap, &instance)
		if err != nil {
			s.logger.Error("failed to decode instance", zap.Error(err), zap.String("key", keys[i]))
			return nil, InstanceRetrievalFailure
		}

		instances[i] = &instance
	}

	return instances, InstanceRetrievalSuccess
}

func (s *instanceService) GetByName(instanceName string) (*models.Instance, InstanceRetrievalResult) {
	var err error
	instanceKey := s.instanceKey(instanceName)

	// retrieve
	cmd := s.redisClient.HGetAll(instanceKey)
	instanceMap, err := cmd.Result()
	if err != nil {
		s.logger.Error("failed to retrieve instance", zap.Error(err), zap.String("instanceName", instanceName))
		return nil, InstanceRetrievalFailure
	}
	if len(instanceMap) == 0 {
		return nil, InstanceRetrievalNotFound
	}

	// decode
	var instance models.Instance
	err = mapstructure.Decode(instanceMap, &instance)
	if err != nil {
		s.logger.Error("failed to decode instance", zap.Error(err), zap.String("instanceName", instanceName))
		return nil, InstanceRetrievalFailure
	}

	return &instance, InstanceRetrievalSuccess
}

func (s *instanceService) doCreate(instance *models.Instance) InstanceCreationResult {
	instanceKey := s.instanceKey(instance.Name)
	instanceMap := structs.Map(instance)

	err := s.redisClient.HMSet(instanceKey, instanceMap).Err()
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
	instance.Status = models.InstanceStatusPending

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
	instanceKey := s.instanceKey(instance.Name)

	value, err := s.redisClient.Del(instanceKey).Result()
	if err != nil {
		s.logger.Error("error while trying to delete instance", zap.String("name", instance.Name), zap.Error(err))
		return InstanceDeletionFailure
	}

	if value == 0 {
		s.logger.Error("instance not found to be deleted", zap.String("name", instance.Name))
		return InstanceDeletionNotFound
	}

	return InstanceDeletionSuccess
}

func (s *instanceService) Delete(instanceName string) InstanceDeletionResult {
	// TODO missing: check if bindings exist before deleting

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

	// delete env vars
	_, _ = s.DelInstanceVars(instance.Name)

	return InstanceDeletionSuccess
}

func (s *instanceService) UpdateStatus(name string, status models.InstanceStatus) InstanceUpdateResult {
	instanceKey := s.instanceKey(name)

	_, err := s.redisClient.HSet(instanceKey, "Status", status).Result()
	if err != nil {
		s.logger.Error("error while trying to update instance", zap.String("name", name), zap.Error(err))
		return InstanceUpdateFailure
	}

	return InstanceUpdateSuccess
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

/*
	===========================================================================
	vars
	===========================================================================
*/
func (s *instanceService) instanceVarsKey(instanceName string) string {
	return fmt.Sprintf("%s:%s", s.instanceVarsKeyPrefix, instanceName)
}

func (s *instanceService) GetInstanceVars(name string) (map[string]string, error) {
	instanceKey := s.instanceVarsKey(name)
	envVars, err := s.redisClient.HGetAll(instanceKey).Result()
	if err != nil {
		s.logger.Error("GetInstanceVars failed", zap.Error(err))
		return nil, err
	}
	return envVars, nil
}

func (s *instanceService) SetInstanceVars(name string, envVars map[string]string) (string, error) {
	instanceKey := s.instanceVarsKey(name)

	// convert string to interface{}
	interfaceMap := make(map[string]interface{}, len(envVars))
	for k, v := range envVars {
		interfaceMap[k] = v
	}

	result, err := s.redisClient.HMSet(instanceKey, interfaceMap).Result()
	if err != nil {
		s.logger.Error("SetInstanceVars failed", zap.Error(err))
		return "", err
	}
	return result, nil
}

func (s *instanceService) DelInstanceVars(name string) (int64, error) {
	instanceKey := s.instanceVarsKey(name)

	result, err := s.redisClient.Del(instanceKey).Result()
	if err != nil {
		s.logger.Error("DelInstanceVars failed", zap.Error(err))
		return 0, err
	}
	return result, nil
}

func NewInstanceService(config *viper.Viper, logger *zap.Logger, redisClient redis.UniversalClient, provisionService ProvisionService) InstanceService {
	instanceKeyPrefix := config.GetString("redis.db.instance.prefix")
	instanceVarsKeyPrefix := config.GetString("redis.db.instance.vars-prefix")

	return &instanceService{
		instanceKeyPrefix:     instanceKeyPrefix,
		instanceVarsKeyPrefix: instanceVarsKeyPrefix,
		logger:                logger,
		provisionService:      provisionService,
		redisClient:           redisClient,
	}
}
