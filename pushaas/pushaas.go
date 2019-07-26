package pushaas

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/ctors"
)

func runApp(logger *zap.Logger, redisClient redis.UniversalClient, router *gin.Engine, config *viper.Viper) error {
	log := logger.Named("runApp")

	err := router.Run(fmt.Sprintf(":%s", config.GetString("server.port")))
	if err != nil {
		log.Error("error on running server", zap.Error(err))
		return err
	}

	return nil
}

func Run() {
	app := fx.New(
		fx.Provide(
			ctors.NewViper,
			ctors.NewLogger,
			ctors.NewRedisClient,

			// routers
			ctors.NewGinRouter,
			ctors.NewRootRouter,
			ctors.NewStaticRouter,
			ctors.NewApiRootRouter,
			ctors.NewAuthRouter,
			ctors.NewResourceRouter,
			ctors.NewBindRouter,

			// services
			ctors.NewInstanceService,
			ctors.NewBindService,
			ctors.NewPlanService,
			ctors.NewProvisionService,

			// provisioners
			ctors.NewProvisioner,
		),
		fx.Invoke(runApp),
	)

	app.Run()
}
