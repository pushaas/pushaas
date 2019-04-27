package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	StaticRouter interface {
		Router
	}

	staticRouter struct{}
)

func (r *staticRouter) getStaticRoot(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Data: gin.H{
			"TODO": "TODO",
		},
	})
}

func (r *staticRouter) SetupRoutes(router gin.IRouter) {
	router.GET("/", r.getStaticRoot)
}

func NewStaticRouter() Router {
	return &staticRouter{}
}
