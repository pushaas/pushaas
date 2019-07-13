package apiV1

import (
	"github.com/gin-gonic/gin"

	"github.com/rafaeleyng/pushaas/pushaas/routers"
)

type (
	AuthRouter interface {
		routers.Router
	}

	authRouter struct {}
)

func (r *authRouter) checkAuth(c *gin.Context) {}

func (r *authRouter) SetupRoutes(router gin.IRouter) {
	router.GET("", r.checkAuth)
}

func NewAuthRouter() AuthRouter {
	return &authRouter{}
}
