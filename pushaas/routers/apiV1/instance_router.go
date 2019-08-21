package apiV1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rafaeleyng/pushaas/pushaas/models"

	"github.com/rafaeleyng/pushaas/pushaas/routers"
	"github.com/rafaeleyng/pushaas/pushaas/services"
)

type (
	InstanceRouter interface {
		routers.Router
	}

	instanceRouter struct {
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

func (r *instanceRouter) getPlansOrInstance(c *gin.Context) {
	name := nameFromPath(c)
	if name == "plans" {
		r.getPlans(c)
	} else if name == "instances" {
		r.getInstances(c)
	} else {
		r.getInstance(c)
	}
}

func (r *instanceRouter) getPlans(c *gin.Context) {
	plans := r.planService.GetAll()
	c.JSON(http.StatusOK, plans)
}

func (r *instanceRouter) getInstances(c *gin.Context) {
	instances, result := r.instanceService.GetAll()

	if result == services.InstanceRetrievalNotFound {
		c.JSON(http.StatusNotFound, models.Error{
			Code:    models.ErrorInstanceRetrievalNotFound,
			Message: "Instance not found",
		})
		return
	}

	if result == services.InstanceRetrievalFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:    models.ErrorInstanceRetrievalFailed,
			Message: "Failed to retrieve instance",
		})
		return
	}

	c.JSON(http.StatusOK, instances)
}

func (r *instanceRouter) getInstance(c *gin.Context) {
	name := nameFromPath(c)
	instance, result := r.instanceService.GetByName(name)

	if result == services.InstanceRetrievalNotFound {
		c.JSON(http.StatusNotFound, models.Error{
			Code:    models.ErrorInstanceRetrievalNotFound,
			Message: "Instance not found",
		})
		return
	}

	if result == services.InstanceRetrievalFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:    models.ErrorInstanceRetrievalFailed,
			Message: "Failed to retrieve instance",
		})
		return
	}

	c.JSON(http.StatusOK, instance)
}

func (r *instanceRouter) postInstance(c *gin.Context) {
	instanceForm := instanceFormFromContext(c)
	result := r.instanceService.Create(instanceForm)

	if result == services.InstanceCreationAlreadyExist {
		c.JSON(http.StatusConflict, models.Error{
			Code:    models.ErrorInstanceCreateAlreadyExists,
			Message: "An instance with this name already exists",
		})
		return
	}

	if result == services.InstanceCreationInvalidData {
		c.JSON(http.StatusBadRequest, models.Error{
			Code:    models.ErrorInstanceCreateInvalidData,
			Message: "Invalid instance data",
		})
		return
	}

	if result == services.InstanceCreationFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:    models.ErrorInstanceCreateFailed,
			Message: "Failed to create instance",
		})
		return
	}

	if result == services.InstanceCreationProvisionFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:    models.ErrorInstanceCreateDispatchProvisionFailed,
			Message: "Instance created, but unable to dispatch provision. Please remove it manually",
		})
		return
	}

	c.Status(http.StatusCreated)
}

func (r *instanceRouter) putInstance(c *gin.Context) {
	// this endpoint is optional, we return 404 to signal Tsuru that is not implemented
	c.Status(http.StatusNotFound)
}

func (r *instanceRouter) deleteInstance(c *gin.Context) {
	name := nameFromPath(c)
	result := r.instanceService.Delete(name)

	if result == services.InstanceDeletionNotFound {
		c.JSON(http.StatusNotFound, models.Error{
			Code:    models.ErrorInstanceDeleteNotFound,
			Message: "Instance not found",
		})
		return
	}

	if result == services.InstanceDeletionFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:    models.ErrorInstanceDeleteFailed,
			Message: "Failed to delete instance",
		})
		return
	}

	if result == services.InstanceDeletionDeprovisionFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:    models.ErrorInstanceDeleteDispatchDeprovisionFailed,
			Message: "Instance deleted, but unable to dispatch deprovision. Please deprovision it manually",
		})
		return
	}

	c.Status(http.StatusOK)
}

func (r *instanceRouter) getInstanceStatus(c *gin.Context) {
	name := nameFromPath(c)
	result := r.instanceService.GetStatusByName(name)

	/*
		when we could not get the actual status
	 */
	if result == services.InstanceStatusNotFound {
		c.JSON(http.StatusNotFound, models.Error{
			Code:    models.ErrorInstanceStatusRetrievalNotFound,
			Message: "Instance not found",
		})
		return
	}

	if result == services.InstanceStatusFailure {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:    models.ErrorInstanceStatusRetrievalFailed,
			Message: "Failed to retrieve instance status",
		})
		return
	}

	/*
		when we've got the actual status
	*/
	if result == services.InstanceStatusFailedStatus {
		c.JSON(http.StatusInternalServerError, models.Error{
			Code:    models.ErrorInstanceStatusInstanceFailed,
			Message: "Instance is in failed status",
		})
		return
	}

	if result == services.InstanceStatusPendingStatus {
		c.Status(http.StatusAccepted)
		return
	}

	c.Status(http.StatusNoContent)
}

func (r *instanceRouter) SetupRoutes(router gin.IRouter) {
	// default / service instance
	router.GET("/:name", r.getPlansOrInstance)

	// service instance
	router.POST("", r.postInstance)
	router.PUT("/:name", r.putInstance)
	router.DELETE("/:name", r.deleteInstance)
	router.GET("/:name/status", r.getInstanceStatus)
}

func NewInstanceRouter(instanceService services.InstanceService, planService services.PlanService) routers.Router {
	return &instanceRouter{
		instanceService: instanceService,
		planService:     planService,
	}
}
