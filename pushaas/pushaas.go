package pushaas

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/ctors"
)

func runApp(logger *zap.Logger, router *gin.Engine, config *viper.Viper) error {
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
			ctors.NewMongodb,

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

			// provisioners
			ctors.NewProvisioner,
		),
		fx.Invoke(runApp),
	)

	app.Run()
}
