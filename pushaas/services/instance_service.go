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
	AppBindAlreadyBound
	AppBindInstancePending
	AppBindInstanceNotFound
	AppBindFailure
)

const (
	AppUnbindSuccess AppUnbindResult = iota
	AppUnbindAlreadyUnbound
	AppUnbindInstanceNotFound
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
	err := instanceForm.Validate()
	if err != nil {
		// TODO
	}
	// TODO validate name is unique
	// TODO validate plan
	// TODO dispatch instance provisioning

	instance := instanceFromInstanceForm(instanceForm)
	instance.Status = models.InstanceStatusPending

	err = s.getCollection().Save(instance)
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

func (s *instanceService) BindApp(name string, bindAppForm *models.BindAppForm) (map[string]string, AppBindResult) {
	//instance, result := s.getByName(name)
	//
	//if result == InstanceRetrievalNotFound {
	//	return nil, AppBindInstanceNotFound
	//}

	//if instance

	envVars := map[string]string{
		"PUSHAAS_ENDPOINT": "TODO-endpoint",
		"PUSHAAS_USERNAME": "TODO-username",
		"PUSHAAS_PASSWORD": "TODO-password",
	}

	return envVars, AppBindSuccess
}

func (s *instanceService) UnbindApp(name string, bindAppForm *models.BindAppForm) AppUnbindResult {
	panic("implement me")
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
