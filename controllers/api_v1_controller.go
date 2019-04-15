package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rafaeleyng/pushaas/business"
)

func handleV1GetServices(c *gin.Context) {
	c.JSON(http.StatusOK, business.Response{
		Data: gin.H{
			"TODO": "TODO",
		},
	})
}

func SetupApiV1Group(router gin.IRouter) {
	router.GET("/services", handleV1GetServices)
}
