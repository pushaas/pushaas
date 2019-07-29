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
	BindAppRetrievalResult int
	BindAppResult          int
	AppUnbindResult        int

	UnitBindResult   int
	UnitUnbindResult int

	BindService interface {
		BindApp(name string, bindAppForm *models.BindAppForm) (map[string]string, BindAppResult)
		UnbindApp(name string, bindAppForm *models.BindAppForm) AppUnbindResult
		BindUnit(name string, bindUnitForm *models.BindUnitForm) UnitBindResult
		UnbindUnit(name string, bindUnitForm *models.BindUnitForm) UnitUnbindResult
	}

	bindService struct {
		bindingsKeyPrefix  string
		unitsHostKeyPrefix string
		instanceService    InstanceService
		logger             *zap.Logger
		redisClient        redis.UniversalClient
	}
)

const (
	BindAppRetrievalSuccess BindAppRetrievalResult = iota
	BindAppRetrievalNotFound
	BindAppRetrievalFailure
)

const (
	BindAppSuccess BindAppResult = iota
	BindAppNotFound
	BindAppAlreadyBound
	BindAppFailure

	BindAppInstancePending
	BindAppInstanceFailed
)

const (
	AppUnbindSuccess AppUnbindResult = iota
	AppUnbindInstanceNotFound
	AppUnbindNotBound
	AppUnbindFailure
)

//const (
//	UnitBindSuccess UnitBindResult = iota
//	UnitBindAlreadyBound
//	UnitBindInstancePending
//	UnitBindInstanceNotFound
//	UnitBindFailure
//)
//
//const (
//	UnitUnbindSuccess UnitUnbindResult = iota
//	UnitUnbindAlreadyUnbound
//	UnitUnbindInstanceNotFound
//	UnitUnbindFailure
//)

func (s *bindService) bindAppKey(instanceName, appName string) string {
	return fmt.Sprintf("%s:%s:%s", s.bindingsKeyPrefix, instanceName, appName)
}

func (s *bindService) getBindApp(instanceName, appName string) (*models.BindApp, BindAppRetrievalResult) {
	var err error
	bindAppKey := s.bindAppKey(instanceName, appName)

	// retrieve
	cmd := s.redisClient.HGetAll(bindAppKey)
	bindAppMap, err := cmd.Result()
	if err != nil {
		s.logger.Error("failed to retrieve bindApp", zap.Error(err), zap.String("bindAppKey", bindAppKey))
		return nil, BindAppRetrievalFailure
	}
	if len(bindAppMap) == 0 {
		return nil, BindAppRetrievalNotFound
	}

	// decode
	var bindApp models.BindApp
	err = mapstructure.Decode(bindAppMap, &bindApp)
	if err != nil {
		s.logger.Error("failed to decode bindApp", zap.Error(err), zap.String("instance", instanceName))
		return nil, BindAppRetrievalFailure
	}

	return &bindApp, BindAppRetrievalSuccess
}

func (s *bindService) doCreateBindApp(instance *models.Instance, bindApp *models.BindApp) BindAppResult {
	bindAppKey := s.bindAppKey(instance.Name, bindApp.AppName)
	bindAppMap := structs.Map(bindApp)

	err := s.redisClient.HMSet(bindAppKey, bindAppMap).Err()
	if err != nil {
		s.logger.Error("failed to create bindApp", zap.Error(err), zap.Any("instance", instance), zap.Any("bindApp", bindApp))
		return BindAppFailure
	}
	return BindAppSuccess
}

func (s *bindService) BindApp(instanceName string, bindAppForm *models.BindAppForm) (map[string]string, BindAppResult) {
	// check instance existence
	instance, resultInstanceGet := s.instanceService.GetByName(instanceName)
	if resultInstanceGet == InstanceRetrievalNotFound {
		s.logger.Error("instance not found for bindApp", zap.String("instanceName", instanceName), zap.Any("bindAppForm", bindAppForm))
		return nil, BindAppNotFound
	}

	// check instance status
	if instance.Status == models.InstanceStatusPending {
		return nil, BindAppInstancePending
	} else if instance.Status == models.InstanceStatusFailed {
		return nil, BindAppInstanceFailed
	}

	// check existing binding
	_, resultBindAppGet := s.getBindApp(instance.Name, bindAppForm.AppName)
	if resultBindAppGet == BindAppRetrievalSuccess {
		s.logger.Error("instance already bound to app", zap.String("instanceName", instanceName), zap.Any("bindAppForm", bindAppForm))
		return nil, BindAppAlreadyBound
	} else if resultBindAppGet == BindAppRetrievalFailure {
		return nil, BindAppFailure
	}

	bindApp := models.BindAppFromForm(bindAppForm)

	// bind
	resultBind := s.doCreateBindApp(instance, bindApp)
	if resultBind != BindAppSuccess {
		return nil, resultBind
	}

	// get instance variables
	// TODO get variables from the real source
	envVars := map[string]string{
		"PUSHAAS_ENDPOINT": "the-endpoint",
		"PUSHAAS_USERNAME": "the-username",
		"PUSHAAS_PASSWORD": "the-password",
	}

	return envVars, BindAppSuccess
}

func (s *bindService) doDeleteBindApp(instance *models.Instance, bindApp *models.BindApp) AppUnbindResult {
	bindAppKey := s.bindAppKey(instance.Name, bindApp.AppName)

	value, err := s.redisClient.Del(bindAppKey).Result()
	if err != nil {
		s.logger.Error("failed to delete bindApp", zap.Error(err), zap.Any("instance", instance), zap.Any("bindApp", bindApp))
		return AppUnbindFailure
	}

	if value == 0 {
		s.logger.Error("bindApp not found to be deleted", zap.String("name", instance.Name))
		return AppUnbindNotBound
	}

	return AppUnbindSuccess
}

func (s *bindService) UnbindApp(instanceName string, bindAppForm *models.BindAppForm) AppUnbindResult {
	// check instance existence
	instance, resultInstanceGet := s.instanceService.GetByName(instanceName)
	if resultInstanceGet == InstanceRetrievalNotFound {
		s.logger.Error("instance not found for binding", zap.String("instanceName", instanceName), zap.Any("bindAppForm", bindAppForm))
		return AppUnbindInstanceNotFound
	}

	// check existing binding
	_, resultBindAppGet := s.getBindApp(instance.Name, bindAppForm.AppName)
	if resultBindAppGet == BindAppRetrievalNotFound {
		s.logger.Error("instance not bound to app", zap.String("instanceName", instanceName), zap.Any("bindAppForm", bindAppForm))
		return AppUnbindNotBound
	} else if resultBindAppGet == BindAppRetrievalFailure {
		return AppUnbindFailure
	}

	bindApp := models.BindAppFromForm(bindAppForm)

	// unbind
	return s.doDeleteBindApp(instance, bindApp)
}

func (s *bindService) BindUnit(name string, bindUnitForm *models.BindUnitForm) UnitBindResult {
	// TODO implement
	panic("implement me")
}

func (s *bindService) UnbindUnit(name string, bindUnitForm *models.BindUnitForm) UnitUnbindResult {
	// TODO implement
	panic("implement me")
}

func NewBindService(config *viper.Viper, logger *zap.Logger, redisClient redis.UniversalClient, instanceService InstanceService) BindService {
	bindingsKeyPrefix := config.GetString("redis.db.bindings.prefix")
	unitsHostKeyPrefix := config.GetString("redis.db.units-host.prefix")

	return &bindService{
		bindingsKeyPrefix:  bindingsKeyPrefix,
		unitsHostKeyPrefix: unitsHostKeyPrefix,
		instanceService:    instanceService,
		logger:             logger,
		redisClient:        redisClient,
	}
}
