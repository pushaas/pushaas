package models

import (
	"github.com/go-bongo/bongo"
	"gopkg.in/mgo.v2/bson"

	"github.com/rafaeleyng/pushaas/pushaas/mongodb"
)

type Instance struct {
	bongo.DocumentBase `bson:",inline"`
	Description        string `json:"description"`
}

func getCollection() *bongo.Collection {
	return mongodb.GetConnection().Collection("instances")
}

func InstanceSave(instance *Instance) (Instance, error) {
	result := *instance
	err := getCollection().Save(&result)
	if err != nil {
		return Instance{}, err
	}
	return result, nil
}

func InstanceGet(id string) (Instance, error) {
	objectId := bson.ObjectIdHex(id)
	result := Instance{}
	err := getCollection().FindById(objectId, &result)

	if err != nil {
		return Instance{}, err
	}

	return result, nil
}

func InstanceGetAll() ([]Instance, error) {
	var result []Instance

	found := getCollection().Find(bson.M{})
	// TODO paginate - https://github.com/go-bongo/bongo#find
	//found.Paginate()

	instance := Instance{}

	for found.Next(&instance) {
		result = append(result, instance)
	}

	return result, nil
}

func InstanceDelete(id string) error {
	objectId := bson.ObjectIdHex(id)

	query := bson.M{"_id": objectId}
	err := getCollection().DeleteOne(query)
	return err
}
