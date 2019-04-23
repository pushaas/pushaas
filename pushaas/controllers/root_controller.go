package controllers

import "github.com/gin-gonic/gin"

func SetupRootRoutes(router gin.IRouter) {
	apiGroup := router.Group("/api")
	staticGroup := router.Group("/static")

	SetupApiRoutes(apiGroup)
	SetupStaticRoutes(staticGroup)
}
