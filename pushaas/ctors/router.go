package ctors

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	v1 "github.com/rafaeleyng/pushaas/pushaas/routers/v1"

	"github.com/rafaeleyng/pushaas/pushaas/routers"

	"github.com/rafaeleyng/pushaas/pushaas/services"
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

func NewRouter(config *viper.Viper, logger *zap.Logger, instanceService services.InstanceService) *gin.Engine {
	r := gin.Default()

	g(r, "/static", func(r gin.IRouter) {
		routers.StaticRouter(r)
	})

	g(r, "/api", func(r gin.IRouter) {
		r.Use(getBasicAuthMiddleware(config, logger))

		g(r, "/", func(r gin.IRouter) {
			routers.ApiRootRouter(r)
		})

		g(r, "/v1", func(r gin.IRouter) {
			g(r, "/instances", func(r gin.IRouter) {
				v1.InstanceRouter(r, logger, instanceService)
			})
		})
	})

	return r
}
