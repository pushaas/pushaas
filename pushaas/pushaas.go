package pushaas

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/pushaas/pushaas/pushaas/ctors"
	"github.com/pushaas/pushaas/pushaas/workers"
)

func runApp(
	logger *zap.Logger,
	router *gin.Engine,
	config *viper.Viper,
	machineryWorker workers.MachineryWorker,
) error {
	log := logger.Named("runApp")

	machineryWorker.DispatchWorker()

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
			ctors.NewMachineryServer,

			// routers
			ctors.NewGinRouter,
			ctors.NewRootRouter,
			ctors.NewStaticRouter,
			ctors.NewApiRootRouter,
			ctors.NewAuthRouter,
			ctors.NewInstanceRouter,
			ctors.NewBindRouter,

			// services
			ctors.NewInstanceService,
			ctors.NewBindService,
			ctors.NewPlanService,
			ctors.NewProvisionService,

			// provisioners
			ctors.NewPushServiceProvisioner,

			// provisioner - ecs
			ctors.NewEcsProvisionerConfig,
			ctors.NewEcsPushRedisProvisioner,
			ctors.NewEcsPushStreamProvisioner,
			ctors.NewEcsPushApiProvisioner,

			// workers
			ctors.NewInstanceWorker,
			ctors.NewProvisionWorker,
			ctors.NewMachineryWorker,
		),
		fx.Invoke(runApp),
	)

	app.Run()
}
