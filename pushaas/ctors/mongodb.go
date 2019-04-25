package ctors

import (
	"github.com/go-bongo/bongo"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func NewMongodb(config *viper.Viper, logger *zap.Logger) (*bongo.Connection, error) {
	connectionString := config.GetString("mongodb.connection_string")
	database := config.GetString("mongodb.database")

	options := &bongo.Config{
		ConnectionString: connectionString,
		Database:         database,
	}

	logger.Info("initializing mongodb with options", zap.String("ConnectionString", connectionString), zap.String("Database", database))

	return bongo.Connect(options)
}
