package ctors

import (
	"errors"

	"github.com/RichardKnop/machinery/v1"
	machineryConfig "github.com/RichardKnop/machinery/v1/config"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func getRedisUrl(config *viper.Viper) string {
	return config.GetString("redis.url")
}

func getRedisOptions(config *viper.Viper) (*redis.Options, error) {
	url := getRedisUrl(config)
	options, err := redis.ParseURL(url)

	if err != nil {
		return nil, errors.New("failed to parse redis URL")
	}

	if options.Addr == "" {
		return nil, errors.New("redis URL is required")
	}

	return options, nil
}

func NewRedisClient(config *viper.Viper, logger *zap.Logger) (redis.UniversalClient, error) {
	log := logger.Named("redisClient")

	options, err := getRedisOptions(config)
	if err != nil {
		log.Error("failed to init redis options", zap.Error(err))
		return nil, err
	}

	client := redis.NewClient(options)
	return client, nil
}

func NewMachineryServer(config *viper.Viper, logger *zap.Logger) (*machinery.Server, error) {
	log := logger.Named("machinery")
	url := getRedisUrl(config)

	var cnf = &machineryConfig.Config{
		Broker:             url,
		DefaultQueue:       "machinery_tasks",
		ResultBackend:      url,
	}

	server, err := machinery.NewServer(cnf)
	if err != nil {
		log.Error("failed to init machinery server", zap.Error(err))
		return nil, err
	}

	return server, nil
}
