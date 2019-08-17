package pushaas

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/ctors"
	"github.com/rafaeleyng/pushaas/pushaas/models"
	"github.com/rafaeleyng/pushaas/pushaas/provisioners"
)

func runApp(
	logger *zap.Logger,
	router *gin.Engine,
	config *viper.Viper,
	//provisionWorker workers.ProvisionWorker,
	provisioner provisioners.PushServiceProvisioner,
) error {
	log := logger.Named("runApp")

	//provisionWorker.DispatchWorker()

	instance := &models.Instance{
		Name:   "instance-57",
	}
	fmt.Println("#### testando instance", instance)
	provisioner.Provision(instance)
	//provisioner.Deprovision(instance)
	//provisioner.CleanupServices()

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
			ctors.NewProvisionWorker,
		),
		fx.Invoke(runApp),
	)

	app.Run()
}
