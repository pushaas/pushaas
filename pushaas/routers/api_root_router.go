package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	apiRootRouter struct{}

	serviceStatus struct {
		Service string `json:"service"`
		Status  string `json:"status"`
	}
)

func (r *apiRootRouter) getApiRoot(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Data: gin.H{
			"app": "pushaas",
		},
	})
}

func (r *apiRootRouter) getApiHealthcheck(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Data: gin.H{
			"services": []serviceStatus{
				{
					Service: "app",
					Status:  "working",
				},
			},
		},
	})
}

func ApiRootRouter(router gin.IRouter) Router {
	r := &apiRootRouter{}

	router.GET("/", r.getApiRoot)
	router.GET("/healthcheck", r.getApiHealthcheck)

	return r
}
