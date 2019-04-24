package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rafaeleyng/pushaas/pushaas/business"
)

type (
	staticRouter struct{}
)

func (r *staticRouter) getStaticRoot(c *gin.Context) {
	c.JSON(http.StatusOK, business.Response{
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
