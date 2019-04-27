package ctors

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/rafaeleyng/pushaas/pushaas/routers/v1"
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
	resourcesRouter v1.ResourcesRouter,
) *gin.Engine {
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

func NewResourcesRouter(instanceService services.InstanceService, planService services.PlanService) v1.ResourcesRouter {
	return v1.NewResourcesRouter(instanceService, planService)
}

func NewApiRootRouter() routers.ApiRootRouter {
	return routers.NewApiRootRouter()
}

func NewStaticRouter() routers.StaticRouter {
	return routers.NewStaticRouter()
}
