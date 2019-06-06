package ctors

import (
	"github.com/gin-gonic/gin"
	"github.com/rafaeleyng/pushaas/pushaas/routers/apiV1"
	"github.com/rafaeleyng/pushaas/pushaas/services"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/routers"
)

func g(router gin.IRouter, path string, groupFn func(r gin.IRouter)) {
	groupFn(router.Group(path))
}

func getBasicAuthMiddleware(config *viper.Viper, logger *zap.Logger) gin.HandlerFunc {
	user := config.GetString("api.basic_auth_user")
	password := config.GetString("api.basic_auth_password")

	logger.Debug("configuring basic auth middleware", zap.String("user", user), zap.String("password", password))

	return gin.BasicAuth(gin.Accounts{
		user: password,
	})
}

func NewRouter(
	config *viper.Viper,
	logger *zap.Logger,
	staticRouter routers.StaticRouter,
	apiRootRouter routers.ApiRootRouter,
	resourcesRouter apiV1.ResourcesRouter,
) *gin.Engine {
	envConfig := config.Get("env")
	if envConfig == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	g(r, "/static", func(r gin.IRouter) {
		staticRouter.SetupRoutes(r)
	})

	g(r, "/api", func(r gin.IRouter) {
		r.Use(getBasicAuthMiddleware(config, logger))

		g(r, "/", func(r gin.IRouter) {
			apiRootRouter.SetupRoutes(r)
		})

		g(r, "/v1", func(r gin.IRouter) {
			g(r, "/resources", func(r gin.IRouter) {
				resourcesRouter.SetupRoutes(r)
			})
		})
	})

	return r
}

func NewResourcesRouter(instanceService services.InstanceService, planService services.PlanService) apiV1.ResourcesRouter {
	return apiV1.NewResourcesRouter(instanceService, planService)
}

func NewApiRootRouter() routers.ApiRootRouter {
	return routers.NewApiRootRouter()
}

func NewStaticRouter() routers.StaticRouter {
	return routers.NewStaticRouter()
}
