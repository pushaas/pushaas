package apiV1

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/pushaas/pushaas/pushaas/models"

	"github.com/pushaas/pushaas/pushaas/routers"
	"github.com/pushaas/pushaas/pushaas/services"
)

type (
	BindRouter interface {
		routers.Router
	}

	bindRouter struct {
		bindService services.BindService
	}
)

func bindAppFormFromPostContext(c *gin.Context) *models.BindAppForm {
	appHost := c.PostForm("app-host")
	appName := c.PostForm("app-name")

	return &models.BindAppForm{
		AppHost: appHost,
		AppName: appName,
	}
}

func (r *bindRouter) postBindApp(c *gin.Context) {
	name := nameFromPath(c)
	bindAppForm := bindAppFormFromPostContext(c)
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

func bindAppFormFromDeleteContext(c *gin.Context) *models.BindAppForm {
	vs, _ := routers.ParseBody(c)
	appHost := vs["app-host"][0]
	appName := vs["app-name"][0]

	return &models.BindAppForm{
		AppHost: appHost,
		AppName: appName,
	}
}

func (r *bindRouter) deleteBindApp(c *gin.Context) {
	name := nameFromPath(c)
	bindAppForm := bindAppFormFromDeleteContext(c)
	result := r.bindService.UnbindApp(name, bindAppForm)

	if result == services.UnbindAppInstanceNotFound {
		c.JSON(http.StatusNotFound, models.Error{
			Code: models.ErrorUnbindAppNotFound,
			Message: "Instance not found",
		})
		return
	}

	if result == services.UnbindAppNotBound {
		c.JSON(http.StatusNotFound, models.Error{
			Code: models.ErrorUnbindAppNotBound,
			Message: "Instance is not bound to app",
		})
		return
	}

	if result == services.UnbindAppFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code: models.ErrorUnbindAppFailed,
			Message: "Failed to unbind instance from app",
		})
		return
	}

	c.Status(http.StatusOK)
}

func bindUnitFormFromPostContext(c *gin.Context) *models.BindUnitForm {
	appHost := c.PostForm("app-host")
	appName := c.PostForm("app-name")
	unitHost := c.PostForm("unit-host")

	return &models.BindUnitForm{
		AppHost:  appHost,
		AppName:  appName,
		UnitHost: unitHost,
	}
}

func (r *bindRouter) postUnitBind(c *gin.Context) {
	name := nameFromPath(c)
	bindUnitForm := bindUnitFormFromPostContext(c)
	envVars, result := r.bindService.BindUnit(name, bindUnitForm)

	if result == services.BindUnitFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code: models.ErrorBindUnitFailed,
			Message: "Failed to bind unit to service",
		})
		return
	}

	if result == services.BindUnitAppNotBound {
		c.JSON(http.StatusNotFound, models.Error{
			Code: models.ErrorBindUnitAppNotBound,
			Message: "App is not bound to service, could not bind unit",
		})
		return
	}

	if result == services.BindUnitAlreadyBound {
		c.JSON(http.StatusBadRequest, models.Error{
			Code: models.ErrorBindUnitAlreadyBound,
			Message: "Unit is already bound to service",
		})
		return
	}

	c.JSON(http.StatusCreated, envVars)
}

func bindUnitFormFromDeleteContext(c *gin.Context) *models.BindUnitForm {
	vs, _ := routers.ParseBody(c)

	appHost := vs["app-host"][0]
	unitHost := vs["unit-host"][0]

	// on Tsuru docs it says this will be send accordingly to the "else" format, but does not seem to be the case
	var appName string
	if len(vs["app-name"]) == 0 {
		appName = strings.Split(appHost, ".")[0]
	} else {
		appName = vs["app-name"][0]
	}

	return &models.BindUnitForm{
		AppHost: appHost,
		AppName: appName,
		UnitHost: unitHost,
	}
}

func (r *bindRouter) deleteUnitBind(c *gin.Context) {
	name := nameFromPath(c)
	bindUnitForm := bindUnitFormFromDeleteContext(c)
	result := r.bindService.UnbindUnit(name, bindUnitForm)

	if result == services.UnbindUnitFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code: models.ErrorUnbindUnitFailed,
			Message: "Failed to unbind unit from service",
		})
		return
	}

	if result == services.UnbindUnitAppNotBound {
		c.JSON(http.StatusNotFound, models.Error{
			Code: models.ErrorUnbindUnitAppNotBound,
			Message: "App is not bound to service, could not unbind unit",
		})
		return
	}

	if result == services.UnbindUnitNotBound {
		c.JSON(http.StatusNotFound, models.Error{
			Code: models.ErrorUnbindUnitNotBound,
			Message: "Unit is not bound to service",
		})
		return
	}

	c.Status(http.StatusOK)
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
