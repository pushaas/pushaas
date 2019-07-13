package ctors

import (
	"github.com/go-bongo/bongo"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func NewMongodb(config *viper.Viper, logger *zap.Logger) (*bongo.Connection, error) {
	url := config.GetString("mongodb.url")
	database := config.GetString("mongodb.database")

	options := &bongo.Config{
		ConnectionString: url,
		Database:         database,
	}

	logger.Info("initializing mongodb with options", zap.String("url", url), zap.String("database", database))

	return bongo.Connect(options)
}
