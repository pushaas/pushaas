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
	appBindRetrievalResult int
	AppBindResult          int
	AppUnbindResult        int

	UnitBindResult   int
	UnitUnbindResult int

	BindService interface {
		BindApp(name string, bindAppForm *models.BindAppForm) (map[string]string, AppBindResult)
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
	appBindRetrievalSuccess appBindRetrievalResult = iota
	appBindRetrievalNotFound
	appBindRetrievalFailure
)

const (
	AppBindSuccess AppBindResult = iota
	AppBindNotFound
	AppBindAlreadyBound
	AppBindFailure

	AppBindInstancePending
	AppBindInstanceFailed
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

func (s *bindService) appBindKey(instanceName, appName string) string {
	return fmt.Sprintf("%s:%s:%s", s.bindingsKeyPrefix, instanceName, appName)
}

func (s *bindService) getAppBind(instanceName, appName string) (*models.AppBind, appBindRetrievalResult) {
	var err error
	appBindKey := s.appBindKey(instanceName, appName)

	// retrieve
	cmd := s.redisClient.HGetAll(appBindKey)
	appBindMap, err := cmd.Result()
	if err != nil {
		s.logger.Error("failed to retrieve appBind", zap.Error(err), zap.String("appBindKey", appBindKey))
		return nil, appBindRetrievalFailure
	}
	if len(appBindMap) == 0 {
		return nil, appBindRetrievalNotFound
	}

	// decode
	var appBind models.AppBind
	err = mapstructure.Decode(appBindMap, &appBind)
	if err != nil {
		s.logger.Error("failed to decode appBind", zap.Error(err), zap.String("instance", instanceName))
		return nil, appBindRetrievalFailure
	}

	return &appBind, appBindRetrievalSuccess
}

func (s *bindService) doCreateAppBind(instance *models.Instance, appBind *models.AppBind) AppBindResult {
	appBindKey := s.appBindKey(instance.Name, appBind.AppName)
	appBindMap := structs.Map(appBind)

	err := s.redisClient.HMSet(appBindKey, appBindMap).Err()
	if err != nil {
		s.logger.Error("failed to create appBind", zap.Error(err), zap.Any("instance", instance), zap.Any("appBind", appBind))
		return AppBindFailure
	}
	return AppBindSuccess
}

func (s *bindService) BindApp(instanceName string, bindAppForm *models.BindAppForm) (map[string]string, AppBindResult) {
	// check instance existence
	instance, resultInstanceGet := s.instanceService.GetByName(instanceName)
	if resultInstanceGet == InstanceRetrievalNotFound {
		s.logger.Error("instance not found for appBind", zap.String("instanceName", instanceName), zap.Any("bindAppForm", bindAppForm))
		return nil, AppBindNotFound
	}

	// check instance status
	if instance.Status == models.InstanceStatusPending {
		return nil, AppBindInstancePending
	} else if instance.Status == models.InstanceStatusFailed {
		return nil, AppBindInstanceFailed
	}

	// check existing binding
	_, resultAppBindGet := s.getAppBind(instance.Name, bindAppForm.AppName)
	if resultAppBindGet == appBindRetrievalSuccess {
		s.logger.Error("instance already bound to app", zap.String("instanceName", instanceName), zap.Any("bindAppForm", bindAppForm))
		return nil, AppBindAlreadyBound
	} else if resultAppBindGet == appBindRetrievalFailure {
		return nil, AppBindFailure
	}

	appBind := models.AppBindFromForm(bindAppForm)

	// bind
	resultBind := s.doCreateAppBind(instance, appBind)
	if resultBind != AppBindSuccess {
		return nil, resultBind
	}

	// get instance variables
	// TODO get variables from the real source
	envVars := map[string]string{
		"PUSHAAS_ENDPOINT": "TODO-endpoint",
		"PUSHAAS_USERNAME": "TODO-username",
		"PUSHAAS_PASSWORD": "TODO-password",
	}

	return envVars, AppBindSuccess
}

func (s *bindService) doDeleteAppBind(instance *models.Instance, appBind *models.AppBind) AppUnbindResult {
	appBindKey := s.appBindKey(instance.Name, appBind.AppName)

	value, err := s.redisClient.Del(appBindKey).Result()
	if err != nil {
		s.logger.Error("failed to delete appBind", zap.Error(err), zap.Any("instance", instance), zap.Any("appBind", appBind))
		return AppUnbindFailure
	}

	if value == 0 {
		s.logger.Error("appBind not found to be deleted", zap.String("name", instance.Name))
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
	_, resultAppBindGet := s.getAppBind(instance.Name, bindAppForm.AppName)
	if resultAppBindGet == appBindRetrievalNotFound {
		s.logger.Error("instance not bound to app", zap.String("instanceName", instanceName), zap.Any("bindAppForm", bindAppForm))
		return AppUnbindNotBound
	} else if resultAppBindGet == appBindRetrievalFailure {
		return AppUnbindFailure
	}

	appBind := models.AppBindFromForm(bindAppForm)

	// unbind
	return s.doDeleteAppBind(instance, appBind)
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
