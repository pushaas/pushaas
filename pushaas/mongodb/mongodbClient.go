package mongodb

import (
	"github.com/go-bongo/bongo"
)

var connection *bongo.Connection

func GetConnection() *bongo.Connection {
	if connection != nil {
		return connection
	}

	config := &bongo.Config{
		// TODO
		ConnectionString: "mongodb://localhost:27017",
		Database:         "pushaas",
	}

	connection, err := bongo.Connect(config)

	if err != nil {
		panic("######## TODO")
	}

	return connection
}
