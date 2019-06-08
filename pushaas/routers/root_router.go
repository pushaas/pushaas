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

func (r *rootRouter) getRoot(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Data: gin.H{
			"app": "pushaas",
		},
	})
}

func (r *rootRouter) SetupRoutes(router gin.IRouter) {
	router.GET("/", r.getRoot)
}

func NewRootRouter() Router {
	return &rootRouter{}
}
