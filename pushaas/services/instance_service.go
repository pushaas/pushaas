package services

import (
	"github.com/go-bongo/bongo"
	"github.com/rafaeleyng/pushaas/pushaas/models"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

type (
	InstanceCreationResult  int
	InstanceRetrievalResult int
	InstanceDeletionResult  int
	InstanceStatusResult    int

	AppBindResult   int
	AppUnbindResult int

	UnitBindResult   int
	UnitUnbindResult int

	InstanceService interface {
		Create(instanceForm *models.InstanceForm) InstanceCreationResult
		GetByName(name string) (*models.Instance, InstanceRetrievalResult)
		Delete(name string) InstanceDeletionResult
		GetStatusByName(name string) InstanceStatusResult
		BindApp(name string, bindAppForm *models.BindAppForm) (map[string]string, AppBindResult)
		UnbindApp(name string, bindAppForm *models.BindAppForm) AppUnbindResult
		BindUnit(name string, bindUnitForm *models.BindUnitForm) UnitBindResult
		UnbindUnit(name string, bindUnitForm *models.BindUnitForm) UnitUnbindResult
	}

	instanceService struct {
		mongodb *bongo.Connection
		logger  *zap.Logger
	}
)

const (
	InstanceCreationSuccess InstanceCreationResult = iota
	InstanceCreationAlreadyExist
	InstanceCreationInvalidPlan
	InstanceCreationFailure
)

const (
	InstanceRetrievalSuccess InstanceRetrievalResult = iota
	InstanceRetrievalNotFound
)

const (
	InstanceDeletionSuccess InstanceDeletionResult = iota
	InstanceDeletionNotFound
	InstanceDeletionFailure
)

const (
	InstanceStatusRunning InstanceStatusResult = iota
	InstanceStatusPending
	InstanceStatusNotFound
	InstanceStatusFailure
)

const (
	AppBindSuccess AppBindResult = iota
	AppBindInstanceNotFound
	AppBindInstancePending
	AppBindInstanceFailed
	AppBindAlreadyBound
	AppBindFailure
)

const (
	AppUnbindSuccess AppUnbindResult = iota
	AppUnbindInstanceNotFound
	AppUnbindNotBound
	AppUnbindFailure
)

const (
	UnitBindSuccess UnitBindResult = iota
	UnitBindAlreadyBound
	UnitBindInstancePending
	UnitBindInstanceNotFound
	UnitBindFailure
)

const (
	UnitUnbindSuccess UnitUnbindResult = iota
	UnitUnbindAlreadyUnbound
	UnitUnbindInstanceNotFound
	UnitUnbindFailure
)

func instanceFromInstanceForm(instanceForm *models.InstanceForm) *models.Instance {
	return &models.Instance{
		Name: instanceForm.Name,
		Plan: instanceForm.Plan,
		Team: instanceForm.Team,
		User: instanceForm.User,
	}
}

func (s *instanceService) getCollection() *bongo.Collection {
	return s.mongodb.Collection("instances")
}

func (s *instanceService) getByName(name string) (*models.Instance, InstanceRetrievalResult) {
	query := bson.M{"name": name}
	results := s.getCollection().Find(query)
	instance := &models.Instance{}
	ok := results.Next(instance)

	if !ok {
		s.logger.Error("instance not found", zap.String("name", name))
		return &models.Instance{}, InstanceRetrievalNotFound
	}

	return instance, InstanceRetrievalSuccess
}

func (s *instanceService) Create(instanceForm *models.InstanceForm) InstanceCreationResult {
	// TODO dispatch instance provisioning

	_, result := s.getByName(instanceForm.Name)
	if result == InstanceRetrievalSuccess {
		return InstanceCreationAlreadyExist
	}

	validationResult := instanceForm.Validate()
	if validationResult == models.InstanceFormInvalidPlan {
		return InstanceCreationInvalidPlan
	}

	instance := instanceFromInstanceForm(instanceForm)
	instance.Status = models.InstanceStatusPending
	//instance.Status = models.InstanceStatusRunning // TODO!!!!!!!!!!!!

	err := s.getCollection().Save(instance)
	if err != nil {
		s.logger.Error("failed to create instance", zap.Error(err), zap.Any("instance", instance))
		return InstanceCreationFailure
	}

	return InstanceCreationSuccess
}

func (s *instanceService) GetByName(name string) (*models.Instance, InstanceRetrievalResult) {
	return s.getByName(name)
}

func (s *instanceService) Delete(name string) InstanceDeletionResult {
	// TODO dispatch instance de-provisioning

	query := bson.M{"name": name}
	changeInfo, err := s.getCollection().Delete(query)

	if err != nil {
		s.logger.Error("error while trying to delete instance", zap.String("name", name), zap.Error(err))
		return InstanceDeletionFailure
	}

	if changeInfo.Removed == 0 {
		s.logger.Error("instance not found to be deleted", zap.String("name", name))
		return InstanceDeletionNotFound
	}

	return InstanceDeletionSuccess
}

func (s *instanceService) GetStatusByName(name string) InstanceStatusResult {
	instance, result := s.getByName(name)

	if result == InstanceRetrievalNotFound {
		s.logger.Error("instance not found to get status", zap.String("name", name))
		return InstanceStatusNotFound
	}

	if instance.Status == models.InstanceStatusPending {
		return InstanceStatusPending
	}

	if instance.Status == models.InstanceStatusFailed {
		return InstanceStatusFailure
	}

	return InstanceStatusRunning
}

func findAppIndexInBindings(appName string, bindings []models.InstanceBinding) int {
	for i, binding := range bindings {
		if binding.AppName == appName {
			return i
		}
	}
	return -1
}

func findAppInBindings(appName string, bindings []models.InstanceBinding) *models.InstanceBinding {
	i := findAppIndexInBindings(appName, bindings)
	if i == -1 {
		return &models.InstanceBinding{}
	}

	return &bindings[i]
}

func (s *instanceService) BindApp(name string, bindAppForm *models.BindAppForm) (map[string]string, AppBindResult) {
	instance, result := s.getByName(name)

	if result == InstanceRetrievalNotFound {
		s.logger.Error("instance not found for binding", zap.String("name", name), zap.Any("bindAppForm", bindAppForm))
		return nil, AppBindInstanceNotFound
	} else if instance.Status == models.InstanceStatusPending {
		return nil, AppBindInstancePending
	} else if instance.Status == models.InstanceStatusFailed {
		return nil, AppBindInstanceFailed
	}

	instanceBinding := findAppInBindings(bindAppForm.AppName, instance.Bindings)
	if instanceBinding.AppName == bindAppForm.AppName {
		return nil, AppBindAlreadyBound
	}

	instanceBinding = &models.InstanceBinding{
		AppName: bindAppForm.AppName,
		AppHost: bindAppForm.AppHost,
	}
	instance.Bindings = append(instance.Bindings, *instanceBinding)
	err := s.getCollection().Save(instance)
	if err != nil {
		s.logger.Error("failed to bind to instance", zap.String("name", name), zap.Any("bindAppForm", bindAppForm), zap.Any("instance", instance))
		return nil, AppBindFailure
	}

	envVars := map[string]string{
		"PUSHAAS_ENDPOINT": "TODO-endpoint",
		"PUSHAAS_USERNAME": "TODO-username",
		"PUSHAAS_PASSWORD": "TODO-password",
	}

	return envVars, AppBindSuccess
}

func (s *instanceService) UnbindApp(name string, bindAppForm *models.BindAppForm) AppUnbindResult {
	instance, result := s.getByName(name)

	if result == InstanceRetrievalNotFound {
		s.logger.Error("instance not found for unbinding", zap.String("name", name), zap.Any("bindAppForm", bindAppForm))
		return AppUnbindInstanceNotFound
	}

	s.logger.Debug("###### ei rafael", zap.Any("name", name), zap.Any("bindAppForm", bindAppForm), zap.Any("instance", instance))
	i := findAppIndexInBindings(bindAppForm.AppName, instance.Bindings)
	if i == -1 {
		s.logger.Error("instance is not bound to app", zap.String("name", name), zap.Any("bindAppForm", bindAppForm))
		return AppUnbindNotBound
	}

	instance.Bindings = append(instance.Bindings[:i], instance.Bindings[i+1:]...)
	err := s.getCollection().Save(instance)
	if err != nil {
		s.logger.Error("failed to unbind to instance", zap.String("name", name), zap.Any("bindAppForm", bindAppForm), zap.Any("instance", instance))
		return AppUnbindFailure
	}

	return AppUnbindSuccess
}

func (s *instanceService) BindUnit(name string, bindUnitForm *models.BindUnitForm) UnitBindResult {
	panic("implement me")
}

func (s *instanceService) UnbindUnit(name string, bindUnitForm *models.BindUnitForm) UnitUnbindResult {
	panic("implement me")
}

func NewInstanceService(logger *zap.Logger, mongodb *bongo.Connection) InstanceService {
	return &instanceService{
		logger:  logger,
		mongodb: mongodb,
	}
}
