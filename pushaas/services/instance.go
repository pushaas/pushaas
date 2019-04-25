package services

import (
	"github.com/go-bongo/bongo"
	"github.com/rafaeleyng/pushaas/pushaas/models"
	"gopkg.in/mgo.v2/bson"
)

type (
	InstanceService interface {
		Delete(id string) error
		Get(id string) (models.Instance, error)
		GetAll() ([]models.Instance, error)
		Save(instance *models.Instance) (models.Instance, error)
	}

	instanceService struct {
		mongodb *bongo.Connection
	}
)

func (s *instanceService) getCollection() *bongo.Collection {
	return s.mongodb.Collection("instances")
}

func (s *instanceService) Save(instance *models.Instance) (models.Instance, error) {
	result := *instance
	err := s.getCollection().Save(&result)
	if err != nil {
		return models.Instance{}, err
	}
	return result, nil
}

func (s *instanceService) Get(id string) (models.Instance, error) {
	objectId := bson.ObjectIdHex(id)
	result := models.Instance{}
	err := s.getCollection().FindById(objectId, &result)

	if err != nil {
		return models.Instance{}, err
	}

	return result, nil
}

func (s *instanceService) GetAll() ([]models.Instance, error) {
	var result []models.Instance

	found := s.getCollection().Find(bson.M{})
	// TODO paginate - https://github.com/go-bongo/bongo#find
	//found.Paginate()

	instance := models.Instance{}

	for found.Next(&instance) {
		result = append(result, instance)
	}

	return result, nil
}

func (s *instanceService) Delete(id string) error {
	objectId := bson.ObjectIdHex(id)

	query := bson.M{"_id": objectId}
	err := s.getCollection().DeleteOne(query)
	return err
}

func NewInstanceService(mongodb *bongo.Connection) InstanceService {
	return &instanceService{
		mongodb: mongodb,
	}
}
