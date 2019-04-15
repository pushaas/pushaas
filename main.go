package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rafaeleyng/pushaas/controllers"
)

func setupRouter() *gin.Engine {
	router := gin.Default()
	controllers.SetupRootRoutes(router)
	return router
}

func main() {
	router := setupRouter()
	router.Run(":9000")
}
