package routers

import "github.com/gin-gonic/gin"

type (
	Router interface {
		SetupRoutes(router gin.IRouter)
	}

	Response struct {
		Data  interface{} `json:"data"`
		Error string      `json:"error,omitempty"`
	}
)
