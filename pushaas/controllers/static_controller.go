package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rafaeleyng/pushaas/pushaas/business"
)

func handleGeStaticRoot(c *gin.Context) {
	c.JSON(http.StatusOK, business.Response{
		Data: gin.H{
			"TODO": "TODO",
		},
	})
}

func SetupStaticRoutes(router gin.IRouter) {
	router.GET("/", handleGeStaticRoot)
}
