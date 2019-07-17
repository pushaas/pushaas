package apiV1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rafaeleyng/pushaas/pushaas/models"

	"github.com/rafaeleyng/pushaas/pushaas/routers"
	"github.com/rafaeleyng/pushaas/pushaas/services"
)

type (
	ResourceRouter interface {
		routers.Router
	}

	resourceRouter struct {
		instanceService services.InstanceService
		planService     services.PlanService
	}
)

func nameFromPath(c *gin.Context) string {
	return c.Param("name")
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

func (r *resourceRouter) getPlansOrInstance(c *gin.Context) {
	name := nameFromPath(c)
	if name == "plans" {
		r.getPlans(c)
	} else {
		r.getInstance(c)
	}
}

func (r *resourceRouter) getPlans(c *gin.Context) {
	plans := r.planService.GetAll()
	c.JSON(http.StatusOK, plans)
}

func (r *resourceRouter) postInstance(c *gin.Context) {
	instanceForm := instanceFormFromContext(c)
	result := r.instanceService.Create(instanceForm)

	if result == services.InstanceCreationFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:    models.ErrorInstanceCreateFailed,
			Message: "failed to create",
		})
		return
	}

	if result == services.InstanceCreationAlreadyExist {
		c.JSON(http.StatusConflict, models.Error{
			Code:    models.ErrorInstanceCreateAlreadyExists,
			Message: "already exists",
		})
		return
	}

	if result == services.InstanceCreationInvalidPlan {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:    models.ErrorInstanceCreateInvalidPlan,
			Message: "invalid plan",
		})
		return
	}

	c.Status(http.StatusCreated)
}

func (r *resourceRouter) getInstance(c *gin.Context) {
	name := nameFromPath(c)
	instance, result := r.instanceService.GetByName(name)

	if result == services.InstanceRetrievalNotFound {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, []models.Instance{*instance})
}

func (r *resourceRouter) putInstance(c *gin.Context) {
	// this endpoint is optional, we return 404 to signal that to Tsuru
	c.Status(http.StatusNotFound)
}

func (r *resourceRouter) deleteInstance(c *gin.Context) {
	name := nameFromPath(c)
	result := r.instanceService.Delete(name)

	if result == services.InstanceDeletionFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:    models.ErrorInstanceDeleteFailed,
			Message: "failed to delete",
		})
		return
	}

	if result == services.InstanceDeletionNotFound {
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

func (r *resourceRouter) getInstanceStatus(c *gin.Context) {
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

func (r *resourceRouter) SetupRoutes(router gin.IRouter) {
	// default / service instance
	router.GET("/:name", r.getPlansOrInstance)

	// service instance
	router.POST("", r.postInstance)
	router.PUT("/:name", r.putInstance)
	router.DELETE("/:name", r.deleteInstance)
	router.GET("/:name/status", r.getInstanceStatus)
}

func NewResourceRouter(instanceService services.InstanceService, planService services.PlanService) routers.Router {
	return &resourceRouter{
		instanceService: instanceService,
		planService:     planService,
	}
}
