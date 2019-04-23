package pushaas

import (
	"github.com/gin-gonic/gin"

	"github.com/rafaeleyng/pushaas/pushaas/controllers"
)

func setupRouter() *gin.Engine {
	router := gin.Default()
	controllers.SetupRootRoutes(router)
	return router
}

func Run() {
	router := setupRouter()
	router.Run(":9000")
}
