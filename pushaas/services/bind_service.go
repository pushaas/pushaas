package services

import (
	"fmt"

	"github.com/fatih/structs"
	"github.com/go-redis/redis"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pushaas/pushaas/pushaas/models"
)

type (
	BindAppRetrievalResult int
	BindAppResult          int
	UnbindAppResult        int

	BindUnitResult int
	UnbindUnitResult int

	BindService interface {
		BindApp(name string, bindAppForm *models.BindAppForm) (map[string]string, BindAppResult)
		UnbindApp(name string, bindAppForm *models.BindAppForm) UnbindAppResult
		BindUnit(name string, bindUnitForm *models.BindUnitForm) BindUnitResult
		UnbindUnit(name string, bindUnitForm *models.BindUnitForm) UnbindUnitResult
	}

	bindService struct {
		bindAppPrefix   string
		bindUnitPrefix  string
		instanceService InstanceService
		logger          *zap.Logger
		redisClient     redis.UniversalClient
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
	UnbindAppSuccess UnbindAppResult = iota
	UnbindAppInstanceNotFound
	UnbindAppNotBound
	UnbindAppFailure
)

const (
	BindUnitSuccess BindUnitResult = iota
	BindUnitAppNotBound
	BindUnitAlreadyBound
	BindUnitFailure
)

const (
	UnbindUnitSuccess UnbindUnitResult = iota
	UnbindUnitAppNotBound
	UnbindUnitNotBound
	UnbindUnitFailure
)

func (s *bindService) bindAppKey(instanceName, appName string) string {
	return fmt.Sprintf("%s:%s:%s", s.bindAppPrefix, instanceName, appName)
}

func (s *bindService) bindUnitKey(instanceName, appName string) string {
	return fmt.Sprintf("%s:%s:%s", s.bindUnitPrefix, instanceName, appName)
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

func (s *bindService) doBindApp(instance *models.Instance, bindApp *models.BindApp) BindAppResult {
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

	// check binding existence
	_, resultBindAppGet := s.getBindApp(instance.Name, bindAppForm.AppName)
	if resultBindAppGet == BindAppRetrievalSuccess {
		s.logger.Error("instance already bound to app", zap.String("instanceName", instanceName), zap.Any("bindAppForm", bindAppForm))
		return nil, BindAppAlreadyBound
	} else if resultBindAppGet == BindAppRetrievalFailure {
		return nil, BindAppFailure
	}

	bindApp := models.BindAppFromForm(bindAppForm)

	// bind
	resultBind := s.doBindApp(instance, bindApp)
	if resultBind != BindAppSuccess {
		return nil, resultBind
	}

	// get instance variables
	var envVars map[string]string
	envVars, err := s.instanceService.GetInstanceVars(instance.Name)
	if err != nil {
		s.logger.Error("could not retrieve env vars for instance", zap.String("instanceName", instanceName), zap.Any("bindAppForm", bindAppForm), zap.Error(err))
		return nil, BindAppFailure
	}

	return envVars, BindAppSuccess
}

func (s *bindService) doUnbindApp(instance *models.Instance, bindApp *models.BindApp) UnbindAppResult {
	bindAppKey := s.bindAppKey(instance.Name, bindApp.AppName)

	value, err := s.redisClient.Del(bindAppKey).Result()
	if err != nil {
		s.logger.Error("failed to delete bindApp", zap.Error(err), zap.Any("instance", instance), zap.Any("bindApp", bindApp))
		return UnbindAppFailure
	}

	if value == 0 {
		s.logger.Error("bindApp not found to be deleted", zap.String("name", instance.Name))
		return UnbindAppNotBound
	}

	return UnbindAppSuccess
}

func (s *bindService) UnbindApp(instanceName string, bindAppForm *models.BindAppForm) UnbindAppResult {
	// check instance existence
	instance, resultInstanceGet := s.instanceService.GetByName(instanceName)
	if resultInstanceGet == InstanceRetrievalNotFound {
		s.logger.Error("instance not found for binding", zap.String("instanceName", instanceName), zap.Any("bindAppForm", bindAppForm))
		return UnbindAppInstanceNotFound
	}

	// check binding existence
	_, resultBindAppGet := s.getBindApp(instance.Name, bindAppForm.AppName)
	if resultBindAppGet == BindAppRetrievalNotFound {
		s.logger.Error("instance not bound to app", zap.String("instanceName", instanceName), zap.Any("bindAppForm", bindAppForm))
		return UnbindAppNotBound
	} else if resultBindAppGet == BindAppRetrievalFailure {
		return UnbindAppFailure
	}

	bindApp := models.BindAppFromForm(bindAppForm)

	// unbind
	return s.doUnbindApp(instance, bindApp)
}

func (s *bindService) doBindUnit(instanceName string, bindUnitForm *models.BindUnitForm) BindUnitResult {
	bindUnitKey := s.bindUnitKey(instanceName, bindUnitForm.AppName)
	unitHost := bindUnitForm.UnitHost

	added, err := s.redisClient.SAdd(bindUnitKey, unitHost).Result()
	if err != nil {
		s.logger.Error("failed to create bindUnit", zap.Error(err), zap.Any("instanceName", instanceName), zap.Any("bindUnitForm", bindUnitForm))
		return BindUnitFailure
	} else if added == 0 {
		return BindUnitAlreadyBound
	}
	return BindUnitSuccess
}

func (s *bindService) BindUnit(instanceName string, bindUnitForm *models.BindUnitForm) BindUnitResult {
	// check binding existence
	_, resultBindAppGet := s.getBindApp(instanceName, bindUnitForm.AppName)
	if resultBindAppGet == BindAppRetrievalNotFound {
		s.logger.Error("instance not bound to app", zap.String("instanceName", instanceName), zap.Any("bindUnitForm", bindUnitForm))
		return BindUnitAppNotBound
	} else if resultBindAppGet == BindAppRetrievalFailure {
		return BindUnitFailure
	}

	// bind
	return s.doBindUnit(instanceName, bindUnitForm)
}

func (s *bindService) doUnbindUnit(instanceName string, bindUnitForm *models.BindUnitForm) UnbindUnitResult {
	bindUnitKey := s.bindUnitKey(instanceName, bindUnitForm.AppName)
	unitHost := bindUnitForm.UnitHost

	removed, err := s.redisClient.SRem(bindUnitKey, unitHost).Result()
	if err != nil {
		s.logger.Error("failed to remove bindUnit", zap.Error(err), zap.Any("instanceName", instanceName), zap.Any("bindUnitForm", bindUnitForm))
		return UnbindUnitFailure
	} else if removed == 0 {
		return UnbindUnitNotBound
	}
	return UnbindUnitSuccess
}

func (s *bindService) UnbindUnit(instanceName string, bindUnitForm *models.BindUnitForm) UnbindUnitResult {
	// check binding existence
	_, resultBindAppGet := s.getBindApp(instanceName, bindUnitForm.AppName)
	if resultBindAppGet == BindAppRetrievalNotFound {
		s.logger.Error("instance not bound to app", zap.String("instanceName", instanceName), zap.Any("bindUnitForm", bindUnitForm))
		return UnbindUnitAppNotBound
	} else if resultBindAppGet == BindAppRetrievalFailure {
		return UnbindUnitFailure
	}

	// bind
	return s.doUnbindUnit(instanceName, bindUnitForm)
}

func NewBindService(config *viper.Viper, logger *zap.Logger, redisClient redis.UniversalClient, instanceService InstanceService) BindService {
	bindAppPrefix := config.GetString("redis.db.bind-app.prefix")
	bindUnitPrefix := config.GetString("redis.db.bind-unit.prefix")

	return &bindService{
		bindAppPrefix:   bindAppPrefix,
		bindUnitPrefix:  bindUnitPrefix,
		instanceService: instanceService,
		logger:          logger,
		redisClient:     redisClient,
	}
}
