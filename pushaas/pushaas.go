package pushaas

import (
	"github.com/gin-gonic/gin"

	"github.com/rafaeleyng/pushaas/pushaas/config"

	"github.com/rafaeleyng/pushaas/pushaas/controllers"
)

func setupRouter() *gin.Engine {
	router := gin.Default()
	controllers.SetupRootRoutes(router)
	return router
}

func Run() {
	config.SetupConfig() // TODO move this to fx initialization

	router := setupRouter()
	router.Run(":9000")
}
