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

func getNoAuthMiddleware(config *viper.Viper, logger *zap.Logger) gin.HandlerFunc {
	logger.Debug("configuring no auth middleware")

	return func(c *gin.Context) {}
}

func getBasicAuthMiddleware(config *viper.Viper, logger *zap.Logger) gin.HandlerFunc {
	user := config.GetString("api.basic_auth_user")
	password := config.GetString("api.basic_auth_password")

	logger.Debug("configuring basic auth middleware", zap.String("user", user), zap.String("password", password))

	return gin.BasicAuth(gin.Accounts{
		user: password,
	})
}

func getAuthMiddleware(config *viper.Viper, logger *zap.Logger) gin.HandlerFunc {
	if enableAuth := config.GetBool("api.enable_auth"); enableAuth {
		return getBasicAuthMiddleware(config, logger)
	}

	return getNoAuthMiddleware(config, logger)
}

func NewGinRouter(
	config *viper.Viper,
	logger *zap.Logger,
	rootRouter routers.RootRouter,
	staticRouter routers.StaticRouter,
	apiRootRouter routers.ApiRootRouter,
	v1ResourceRouter apiV1.ResourceRouter,
	v1BindRouter apiV1.BindRouter,
) *gin.Engine {
	envConfig := config.Get("env")
	if envConfig == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	baseRouter := gin.Default()

	g(baseRouter, "/", func(r gin.IRouter) {
		rootRouter.SetupRoutes(r)
	})

	g(baseRouter, "/api", func(r gin.IRouter) {
		r.Use(getAuthMiddleware(config, logger))

		g(r, "/", func(r gin.IRouter) {
			apiRootRouter.SetupRoutes(r)
		})

		g(r, "/v1", func(r gin.IRouter) {
			g(r, "/resources", func(r gin.IRouter) {
				v1ResourceRouter.SetupRoutes(r)
				v1BindRouter.SetupRoutes(r)
			})
		})
	})

	g(baseRouter, "/admin", func(r gin.IRouter) {
		staticRouter.SetupRoutes(r)
		staticRouter.SetupClientSideRoutesSupport(baseRouter)
	})

	return baseRouter
}

func NewRootRouter() routers.RootRouter {
	return routers.NewRootRouter()
}

func NewStaticRouter(config *viper.Viper) routers.StaticRouter {
	return routers.NewStaticRouter(config)
}

func NewApiRootRouter() routers.ApiRootRouter {
	return routers.NewApiRootRouter()
}

func NewAuthRouter() apiV1.AuthRouter {
	return apiV1.NewAuthRouter()
}

func NewResourceRouter(instanceService services.InstanceService, planService services.PlanService) apiV1.ResourceRouter {
	return apiV1.NewResourceRouter(instanceService, planService)
}

func NewBindRouter(instanceService services.InstanceService) apiV1.BindRouter {
	return apiV1.NewBindRouter(instanceService)
}
