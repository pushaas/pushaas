package apiV1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rafaeleyng/pushaas/pushaas/models"

	"github.com/rafaeleyng/pushaas/pushaas/routers"
	"github.com/rafaeleyng/pushaas/pushaas/services"
)

type (
	BindRouter interface {
		routers.Router
	}

	bindRouter struct {
		bindService services.BindService
	}
)

func bindAppFormFromContext(c *gin.Context) *models.BindAppForm {
	appHost := c.PostForm("app-host")
	appName := c.PostForm("app-name")

	if appHost == "" && appName == "" {
		vs, _ := routers.ParseBody(c)
		appHost = vs["app-host"][0]
		appName = vs["app-name"][0]
	}

	return &models.BindAppForm{
		AppHost: appHost,
		AppName: appName,
	}
}

func bindUnitFormFromContext(c *gin.Context) *models.BindUnitForm {
	appHost := c.PostForm("app-host")
	appName := c.PostForm("app-name")
	unitHost := c.PostForm("unit-host")

	return &models.BindUnitForm{
		AppHost:  appHost,
		AppName:  appName,
		UnitHost: unitHost,
	}
}

func (r *bindRouter) postBindApp(c *gin.Context) {
	name := nameFromPath(c)
	bindAppForm := bindAppFormFromContext(c)
	envVars, result := r.bindService.BindApp(name, bindAppForm)

	if result == services.BindAppNotFound {
		c.JSON(http.StatusNotFound, models.Error{
			Code: models.ErrorBindAppNotFound,
			Message: "Instance not found",
		})
		return
	}

	if result == services.BindAppInstancePending {
		c.JSON(http.StatusPreconditionFailed, models.Error{
			Code: models.ErrorBindAppInstancePending,
			Message: "Instance is in pending status",
		})
		return
	}

	if result == services.BindAppInstanceFailed {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code: models.ErrorBindAppInstanceFailed,
			Message: "Instance is in failed status",
		})
		return
	}

	if result == services.BindAppAlreadyBound {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code: models.ErrorBindAppAlreadyBound,
			Message: "Instance is already bound to app",
		})
		return
	}

	if result == services.BindAppFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code: models.ErrorBindAppFailed,
			Message: "Failed to bind instance to app",
		})
		return
	}

	c.JSON(http.StatusCreated, envVars)
}

func (r *bindRouter) deleteBindApp(c *gin.Context) {
	name := nameFromPath(c)
	bindAppForm := bindAppFormFromContext(c)
	result := r.bindService.UnbindApp(name, bindAppForm)

	if result == services.AppUnbindInstanceNotFound {
		c.JSON(http.StatusNotFound, models.Error{
			Code: models.ErrorUnbindAppNotFound,
			Message: "Instance not found",
		})
		return
	}

	if result == services.AppUnbindNotBound {
		c.JSON(http.StatusNotFound, models.Error{
			Code: models.ErrorUnbindAppNotBound,
			Message: "Instance is not bound to app",
		})
		return
	}

	if result == services.AppUnbindFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code: models.ErrorUnbindAppFailed,
			Message: "Failed to unbind instance from app",
		})
		return
	}

	c.Status(http.StatusOK)
}

func (r *bindRouter) postUnitBind(c *gin.Context) {
	// TODO implement
	name := nameFromPath(c)
	bindUnitForm := bindUnitFormFromContext(c)
	result := r.bindService.BindUnit(name, bindUnitForm)
	fmt.Println("result", result)
}

func (r *bindRouter) deleteUnitBind(c *gin.Context) {
	// TODO implement
	name := nameFromPath(c)
	bindUnitForm := bindUnitFormFromContext(c)
	result := r.bindService.UnbindUnit(name, bindUnitForm)
	fmt.Println("result", result)
}

func (r *bindRouter) SetupRoutes(router gin.IRouter) {
	// app bind
	router.POST("/:name/bind-app", r.postBindApp)
	router.DELETE("/:name/bind-app", r.deleteBindApp)

	// unit bind
	router.POST("/:name/bind", r.postUnitBind)
	router.DELETE("/:name/bind", r.deleteUnitBind)
}

func NewBindRouter(bindService services.BindService) routers.Router {
	return &bindRouter{
		bindService: bindService,
	}
}
