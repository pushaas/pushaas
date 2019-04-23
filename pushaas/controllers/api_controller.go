package controllers

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/rafaeleyng/pushaas/pushaas/controllers/v1"
)

func SetupApiRoutes(router gin.IRouter) {
	apiRootGroup := router.Group("/")
	apiV1Group := router.Group("/v1")

	SetupApiRootGroup(apiRootGroup)
	v1.SetupApiV1Group(apiV1Group)
}
