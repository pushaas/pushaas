package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	staticRouter struct{}
)

func (r *staticRouter) getStaticRoot(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Data: gin.H{
			"TODO": "TODO",
		},
	})
}

func StaticRouter(router gin.IRouter) Router {
	r := &staticRouter{}

	router.GET("/", r.getStaticRoot)

	return r
}
