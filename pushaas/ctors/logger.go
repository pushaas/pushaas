package ctors

import (
	"github.com/spf13/viper"

	"go.uber.org/zap"
)

func NewLogger(config *viper.Viper) (*zap.Logger, error) {
	envConfig := config.Get("env")

	var logger *zap.Logger
	var err error

	if envConfig == "prod" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	return logger, err
}
