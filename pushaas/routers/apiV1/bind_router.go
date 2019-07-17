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
		instanceService services.InstanceService
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

func (r *bindRouter) postAppBind(c *gin.Context) {
	name := nameFromPath(c)
	bindAppForm := bindAppFormFromContext(c)
	envVars, result := r.instanceService.BindApp(name, bindAppForm)

	if result == services.AppBindInstanceNotFound {
		c.Status(http.StatusNotFound)
		return
	}

	if result == services.AppBindInstancePending {
		c.Status(http.StatusPreconditionFailed)
		return
	}

	if result == services.AppBindInstanceFailed {
		c.JSON(http.StatusInternalServerError, models.Error{
			// TODO add remaining fields
			Message: "instance failed",
		})
		return
	}

	if result == services.AppBindAlreadyBound {
		c.JSON(http.StatusInternalServerError, models.Error{
			// TODO add remaining fields
			Message: "already bound",
		})
		return
	}

	if result == services.AppBindFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			// TODO add remaining fields
			Message: "failure to bind",
		})
		return
	}

	c.JSON(http.StatusCreated, envVars)
}

func (r *bindRouter) deleteAppBind(c *gin.Context) {
	name := nameFromPath(c)
	bindAppForm := bindAppFormFromContext(c)
	result := r.instanceService.UnbindApp(name, bindAppForm)

	if result == services.AppUnbindInstanceNotFound {
		c.Status(http.StatusNotFound)
		return
	}

	if result == services.AppUnbindNotBound {
		c.JSON(http.StatusInternalServerError, models.Error{
			// TODO add remaining fields
			Message: "not bound",
		})
		return
	}

	if result == services.AppUnbindFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			// TODO add remaining fields
			Message: "failure to unbind",
		})
		return
	}

	c.Status(http.StatusOK)
}

func (r *bindRouter) postUnitBind(c *gin.Context) {
	// TODO implement
	name := nameFromPath(c)
	bindUnitForm := bindUnitFormFromContext(c)
	result := r.instanceService.BindUnit(name, bindUnitForm)
	fmt.Println("result", result)
}

func (r *bindRouter) deleteUnitBind(c *gin.Context) {
	// TODO implement
	name := nameFromPath(c)
	bindUnitForm := bindUnitFormFromContext(c)
	result := r.instanceService.UnbindUnit(name, bindUnitForm)
	fmt.Println("result", result)
}

func (r *bindRouter) SetupRoutes(router gin.IRouter) {
	// app bind
	router.POST("/:name/bind-app", r.postAppBind)
	router.DELETE("/:name/bind-app", r.deleteAppBind)

	// unit bind
	router.POST("/:name/bind", r.postUnitBind)
	router.DELETE("/:name/bind", r.deleteUnitBind)
}

func NewBindRouter(instanceService services.InstanceService) routers.Router {
	return &bindRouter{
		instanceService: instanceService,
	}
}
