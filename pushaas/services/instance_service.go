package services

import (
	"github.com/go-bongo/bongo"
	"github.com/rafaeleyng/pushaas/pushaas/models"
	"github.com/rafaeleyng/pushaas/pushaas/provisioners"

	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
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

		GetCollection() *bongo.Collection
	}

	instanceService struct {
		mongodb *bongo.Connection
		logger  *zap.Logger
		provisioner provisioners.Provisioner
	}
)

const (
	InstanceCreationSuccess InstanceCreationResult = iota
	InstanceCreationAlreadyExist
	InstanceCreationInvalidPlan
	InstanceCreationFailure
	InstanceCreationProvisioningFailure
)

const (
	InstanceRetrievalSuccess InstanceRetrievalResult = iota
	InstanceRetrievalNotFound
	InstanceRetrievalFailure // TODO
)

const (
	InstanceDeletionSuccess InstanceDeletionResult = iota
	InstanceDeletionNotFound
	InstanceDeletionFailure
	InstanceDeletionDeprovisioningFailure
)

const (
	InstanceStatusRunning InstanceStatusResult = iota
	InstanceStatusPending
	InstanceStatusNotFound
	InstanceStatusFailure
)

func instanceFromInstanceForm(instanceForm *models.InstanceForm) *models.Instance {
	return &models.Instance{
		Name: instanceForm.Name,
		Plan: instanceForm.Plan,
		Team: instanceForm.Team,
		User: instanceForm.User,
	}
}

func (s *instanceService) GetCollection() *bongo.Collection {
	return s.mongodb.Collection("instances")
}

func (s *instanceService) Create(instanceForm *models.InstanceForm) InstanceCreationResult {
	instanceName := instanceForm.Name

	_, result := s.GetByName(instanceForm.Name)
	if result == InstanceRetrievalSuccess {
		return InstanceCreationAlreadyExist
	}

	validationResult := instanceForm.Validate()
	if validationResult == models.InstanceFormInvalidPlan {
		return InstanceCreationInvalidPlan
	}

	instance := instanceFromInstanceForm(instanceForm)
	instance.Status = models.InstanceStatusPending

	err := s.GetCollection().Save(instance)
	if err != nil {
		s.logger.Error("failed to create instance", zap.Error(err), zap.Any("instance", instance))
		return InstanceCreationFailure
	}

	err = s.provisioner.Provision(instanceName)
	//instance.Status = models.InstanceStatusRunning // TODO!!!!!!!!!!!!
	if err != nil {
		s.logger.Error("failed to provision instance", zap.Error(err), zap.Any("instance", instance))
		// TODO handle this error on router
		return InstanceCreationProvisioningFailure
	}

	return InstanceCreationSuccess
}

func (s *instanceService) GetByName(instanceName string) (*models.Instance, InstanceRetrievalResult) {
	query := bson.M{"name": instanceName}
	results := s.GetCollection().Find(query)
	instance := &models.Instance{}
	ok := results.Next(instance)

	if !ok {
		s.logger.Error("instance not found", zap.String("name", instanceName))
		return &models.Instance{}, InstanceRetrievalNotFound
	}

	return instance, InstanceRetrievalSuccess
}

func (s *instanceService) Delete(instanceName string) InstanceDeletionResult {
	// TODO dispatch instance de-provisioning

	query := bson.M{"name": instanceName}
	changeInfo, err := s.GetCollection().Delete(query)

	if err != nil {
		s.logger.Error("error while trying to delete instance", zap.String("name", instanceName), zap.Error(err))
		return InstanceDeletionFailure
	}

	if changeInfo.Removed == 0 {
		s.logger.Error("instance not found to be deleted", zap.String("name", instanceName))
		return InstanceDeletionNotFound
	}

	err = s.provisioner.Deprovision(instanceName)
	if err != nil {
		// TODO handle this error on router
		return InstanceDeletionDeprovisioningFailure
	}

	return InstanceDeletionSuccess
}

func (s *instanceService) GetStatusByName(name string) InstanceStatusResult {
	instance, result := s.GetByName(name)

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

func NewInstanceService(logger *zap.Logger, mongodb *bongo.Connection, provisioner provisioners.Provisioner) InstanceService {
	return &instanceService{
		logger:  logger,
		mongodb: mongodb,
		provisioner: provisioner,
	}
}
