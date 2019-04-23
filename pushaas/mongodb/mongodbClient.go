package mongodb

import (
	"fmt"

	"github.com/go-bongo/bongo"
	"github.com/rafaeleyng/pushaas/pushaas/config"
)

var connection *bongo.Connection

func getConfig() *bongo.Config {
	fmt.Println("########################### oi")

	fmt.Println("connection_string", config.Config.GetString("mongodb.connection_string"))
	fmt.Println("database", config.Config.GetString("mongodb.database"))

	return &bongo.Config{
		ConnectionString: config.Config.GetString("mongodb.connection_string"),
		Database:         config.Config.GetString("mongodb.database"),
	}
}

func GetConnection() *bongo.Connection {
	fmt.Println("########################### oi 2")

	if connection != nil {
		return connection
	}

	connection, err := bongo.Connect(getConfig())

	if err != nil {
		panic("######## TODO")
	}

	return connection
}
