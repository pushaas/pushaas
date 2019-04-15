package controllers

import "github.com/gin-gonic/gin"

func SetupApiRoutes(router gin.IRouter) {
	apiRootGroup := router.Group("/")
	apiV1Group := router.Group("/v1")

	SetupApiRootGroup(apiRootGroup)
	SetupApiV1Group(apiV1Group)
}
