package ctors

import (
	"github.com/go-bongo/bongo"
	"github.com/spf13/viper"
)

func NewMongodb(config *viper.Viper) (*bongo.Connection, error) {
	options := &bongo.Config{
		ConnectionString: config.GetString("mongodb.connection_string"),
		Database:         config.GetString("mongodb.database"),
	}

	return bongo.Connect(options)
}
