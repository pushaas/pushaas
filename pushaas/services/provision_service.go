package services

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/models"
)

type (
	DispatchProvisionResult   int
	DispatchDeprovisionResult int

	ProvisionService interface {
		DispatchProvision(*models.Instance) DispatchProvisionResult
		DispatchDeprovision(*models.Instance) DispatchDeprovisionResult
	}

	provisionService struct {
		logger      *zap.Logger
		redisClient redis.UniversalClient
	}
)

const (
	DispatchProvisionResultSuccess DispatchProvisionResult = iota
)

const (
	DispatchDeprovisionResultSuccess DispatchDeprovisionResult = iota
)

func (provisionService) DispatchProvision(*models.Instance) DispatchProvisionResult {
	// TODO
	return DispatchProvisionResultSuccess
}

func (provisionService) DispatchDeprovision(*models.Instance) DispatchDeprovisionResult {
	// TODO
	return DispatchDeprovisionResultSuccess
}

func NewProvisionService(config *viper.Viper, logger *zap.Logger, redisClient redis.UniversalClient) ProvisionService {
	return &provisionService{
		logger:      logger,
		redisClient: redisClient,
	}
}