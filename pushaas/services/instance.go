package services

import (
	"github.com/go-bongo/bongo"
	"gopkg.in/mgo.v2/bson"
)

type (
	InstanceService interface {
		Delete(id string) error
		Get(id string) (Instance, error)
		GetAll() ([]Instance, error)
		Save(instance *Instance) (Instance, error)
	}

	instanceService struct {
		mongodb *bongo.Connection
	}

	Instance struct {
		bongo.DocumentBase `bson:",inline"`
		Description        string `json:"description"`
	}
)

func (s *instanceService) getCollection() *bongo.Collection {
	return s.mongodb.Collection("instances")
}

func (s *instanceService) Save(instance *Instance) (Instance, error) {
	result := *instance
	err := s.getCollection().Save(&result)
	if err != nil {
		return Instance{}, err
	}
	return result, nil
}

func (s *instanceService) Get(id string) (Instance, error) {
	objectId := bson.ObjectIdHex(id)
	result := Instance{}
	err := s.getCollection().FindById(objectId, &result)

	if err != nil {
		return Instance{}, err
	}

	return result, nil
}

func (s *instanceService) GetAll() ([]Instance, error) {
	var result []Instance

	found := s.getCollection().Find(bson.M{})
	// TODO paginate - https://github.com/go-bongo/bongo#find
	//found.Paginate()

	instance := Instance{}

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
