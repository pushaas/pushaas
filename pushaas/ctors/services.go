package ctors

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/services"
)

func NewPlanService() services.PlanService {
	return services.NewPlanService()
}

func NewBindService(config *viper.Viper, logger *zap.Logger, redisClient redis.UniversalClient, instanceService services.InstanceService) services.BindService {
	return services.NewBindService(config, logger, redisClient, instanceService)
}

func NewProvisionService(config *viper.Viper, logger *zap.Logger, redisClient redis.UniversalClient) services.ProvisionService {
	return services.NewProvisionService(config, logger, redisClient)
}

func NewInstanceService(config *viper.Viper, logger *zap.Logger, redisClient redis.UniversalClient, provisionService services.ProvisionService) services.InstanceService {
	return services.NewInstanceService(config, logger, redisClient, provisionService)
}
