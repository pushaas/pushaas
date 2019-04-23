package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rafaeleyng/pushaas/pushaas/business"
)

type serviceStatus struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

func handleGetApiRoot(c *gin.Context) {
	c.JSON(http.StatusOK, business.Response{
		Data: gin.H{
			"app": "pushaas",
		},
	})
}

func handleGetApiHealthcheck(c *gin.Context) {
	c.JSON(http.StatusOK, business.Response{
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

func SetupApiRootGroup(router gin.IRouter) {
	router.GET("/", handleGetApiRoot)
	router.GET("/healthcheck", handleGetApiHealthcheck)
}
