package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rafaeleyng/pushaas/pushaas/models"

	"github.com/rafaeleyng/pushaas/pushaas/routers"
	"github.com/rafaeleyng/pushaas/pushaas/services"
)

type (
	ResourcesRouter interface {
		routers.Router
	}

	resourcesRouter struct {
		instanceService services.InstanceService
		planService     services.PlanService
	}
)

func nameFromPath(c *gin.Context) string {
	return c.Param("name")
}

func bindAppFormFromContext(c *gin.Context) *models.BindAppForm {
	appHost := c.PostForm("app-host")
	appName := c.PostForm("app-name")

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

func instanceFormFromContext(c *gin.Context) *models.InstanceForm {
	name := c.PostForm("name")
	plan := c.PostForm("plan")
	team := c.PostForm("team")
	user := c.PostForm("user")

	return &models.InstanceForm{
		Name: name,
		Plan: plan,
		Team: team,
		User: user,
	}
}

func (r *resourcesRouter) getPlansOrInstance(c *gin.Context) {
	name := nameFromPath(c)
	if name == "plans" {
		r.getPlans(c)
	} else {
		r.getInstance(c)
	}
}

func (r *resourcesRouter) getPlans(c *gin.Context) {
	plans := r.planService.GetAll()
	c.JSON(http.StatusOK, plans)
}

func (r *resourcesRouter) postInstance(c *gin.Context) {
	instanceForm := instanceFormFromContext(c)
	result := r.instanceService.Create(instanceForm)

	if result == services.InstanceCreationFailure {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusCreated)
}

func (r *resourcesRouter) getInstance(c *gin.Context) {
	name := nameFromPath(c)
	instance, result := r.instanceService.GetByName(name)

	if result == services.InstanceRetrievalNotFound {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, []models.Instance{*instance})
}

func (r *resourcesRouter) putInstance(c *gin.Context) {
	// this endpoint is optional, we return 404 to signal that to Tsuru
	c.Status(http.StatusNotFound)
}

func (r *resourcesRouter) deleteInstance(c *gin.Context) {
	name := nameFromPath(c)
	result := r.instanceService.Delete(name)

	if result == services.InstanceDeletionFailure {
		c.Status(http.StatusInternalServerError)
		return
	}

	if result == services.InstanceDeletionNotFound {
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

func (r *resourcesRouter) getInstanceStatus(c *gin.Context) {
	name := nameFromPath(c)
	result := r.instanceService.GetStatusByName(name)

	if result == services.InstanceStatusPending {
		c.Status(http.StatusAccepted)
		return
	}

	if result == services.InstanceStatusNotFound {
		c.Status(http.StatusNotFound)
		return
	}

	if result == services.InstanceStatusFailure {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func (r *resourcesRouter) postAppBind(c *gin.Context) {
	name := nameFromPath(c)
	bindAppForm := bindAppFormFromContext(c)
	envVars, result := r.instanceService.BindApp(name, bindAppForm)
	fmt.Println("result", result, envVars)
}

func (r *resourcesRouter) deleteAppBind(c *gin.Context) {
	name := nameFromPath(c)
	bindAppForm := bindAppFormFromContext(c)
	result := r.instanceService.UnbindApp(name, bindAppForm)
	fmt.Println("result", result)
}

func (r *resourcesRouter) postUnitBind(c *gin.Context) {
	name := nameFromPath(c)
	bindUnitForm := bindUnitFormFromContext(c)
	result := r.instanceService.BindUnit(name, bindUnitForm)
	fmt.Println("result", result)
}

func (r *resourcesRouter) deleteUnitBind(c *gin.Context) {
	name := nameFromPath(c)
	bindUnitForm := bindUnitFormFromContext(c)
	result := r.instanceService.UnbindUnit(name, bindUnitForm)
	fmt.Println("result", result)
}

func (r *resourcesRouter) SetupRoutes(router gin.IRouter) {
	// default / service instance
	router.GET("/:name", r.getPlansOrInstance)

	// service instance
	router.POST("", r.postInstance)
	router.PUT("/:name", r.putInstance)
	router.DELETE("/:name", r.deleteInstance)
	router.GET("/:name/status", r.getInstanceStatus)

	// binding
	router.POST("/:name/bind-app", r.postAppBind)
	router.DELETE("/:name/bind-app", r.deleteAppBind)
	router.POST("/:name/bind", r.postUnitBind)
	router.DELETE("/:name/bind", r.deleteUnitBind)
}

func NewResourcesRouter(instanceService services.InstanceService, planService services.PlanService) routers.Router {
	return &resourcesRouter{
		instanceService: instanceService,
		planService:     planService,
	}
}
