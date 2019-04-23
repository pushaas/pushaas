package v1

import "github.com/gin-gonic/gin"

func SetupApiV1Group(router gin.IRouter) {
	router.GET("/instances", handleGetInstances)
	router.GET("/instances/:id", handleGetInstance)
	router.DELETE("/instances/:id", handleDeleteInstance)
	router.POST("/instances", handlePostInstance)
}
