package ctors

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/rafaeleyng/pushaas/pushaas/routers/v1"

	"github.com/rafaeleyng/pushaas/pushaas/routers"

	"github.com/rafaeleyng/pushaas/pushaas/services"
)

func g(router gin.IRouter, path string, groupFn func(r gin.IRouter)) {
	groupFn(router.Group(path))
}

func NewRouter(logger *zap.Logger, instanceService services.InstanceService) *gin.Engine {
	r := gin.Default()

	g(r, "/static", func(r gin.IRouter) {
		routers.StaticRouter(r)
	})

	g(r, "/api", func(r gin.IRouter) {
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
