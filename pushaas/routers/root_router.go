package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	RootRouter interface {
		Router
	}

	rootRouter struct{}
)

func (r *rootRouter) SetupRoutes(router gin.IRouter) {
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin")
	})
}

func NewRootRouter() RootRouter {
	return &rootRouter{}
}
