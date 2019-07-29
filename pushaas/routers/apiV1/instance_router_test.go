package apiV1_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rafaeleyng/pushaas/pushaas/mocks"
	"github.com/rafaeleyng/pushaas/pushaas/models"
	"github.com/rafaeleyng/pushaas/pushaas/routers/apiV1"
	"github.com/rafaeleyng/pushaas/pushaas/services"
)

var _ = Describe("InstanceRouter", func() {
	instanceName := "instance-1"
	instanceForm := &models.InstanceForm{
		Name: instanceName,
		Plan: "plan",
		Team: "team",
		User: "user",
	}

	prepareGinRouter := func(instanceService services.InstanceService, planService services.PlanService) *gin.Engine {
		ginRouter := gin.New()
		router := apiV1.NewInstanceRouter(instanceService, planService)
		router.SetupRoutes(ginRouter)
		return ginRouter
	}

	bodyToError := func(recorder *httptest.ResponseRecorder) *models.Error {
		var body *models.Error
		_ = json.Unmarshal([]byte(recorder.Body.String()), &body)
		return body
	}

	bodyToInstance := func(recorder *httptest.ResponseRecorder) *models.Instance {
		var body *models.Instance
		_ = json.Unmarshal([]byte(recorder.Body.String()), &body)
		return body
	}

	bodyToPlans := func(recorder *httptest.ResponseRecorder) []models.Plan {
		var body []models.Plan
		_ = json.Unmarshal([]byte(recorder.Body.String()), &body)
		return body
	}

	_ = Describe("GET plans", func() {
		_ = It("sends all the plans", func() {
			// arrange
			expected := []models.Plan{
				{
					Name:        models.PlanSmall,
					Description: "The only plan",
				},
			}
			instanceService := &mocks.InstanceServiceMock{}
			planService := &mocks.PlanServiceMock{
				GetAllFunc: func() []models.Plan {
					return expected
				},
			}
			ginRouter := prepareGinRouter(instanceService, planService)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/plans", nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToPlans(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Code).To(Equal(200))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(0))
			Expect(planService.GetAllCalls()).To(HaveLen(1))
		})
	})

	_ = Describe("PUT instance", func() {
		_ = It("always returns 404", func() {
			// arrange
			instanceService := &mocks.InstanceServiceMock{}
			planService := &mocks.PlanServiceMock{}
			ginRouter := prepareGinRouter(instanceService, planService)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", fmt.Sprintf("/%s", instanceName), nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(recorder.Code).To(Equal(404))
		})
	})

	_ = Describe("GET instance", func() {
		_ = It("returns the instance if found", func() {
			// arrange
			expected := &models.Instance{
				Name: instanceName,
			}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (instance *models.Instance, result services.InstanceRetrievalResult) {
					return expected, services.InstanceRetrievalSuccess
				},
			}
			planService := &mocks.PlanServiceMock{}
			ginRouter := prepareGinRouter(instanceService, planService)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/%s", instanceName), nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToInstance(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Body.String()).To(Equal(`{"name":"instance-1","plan":"","team":"","user":"","status":""}`))
			Expect(recorder.Code).To(Equal(200))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
			Expect(planService.GetAllCalls()).To(HaveLen(0))
		})

		_ = It("returns error for instance not found", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorInstanceRetrievalNotFound,
				Message: "Instance not found",
			}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (instance *models.Instance, result services.InstanceRetrievalResult) {
					return nil, services.InstanceRetrievalNotFound
				},
			}
			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/%s", instanceName), nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToError(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Code).To(Equal(404))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
		})

		_ = It("returns error when failure occurs", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorInstanceRetrievalFailed,
				Message: "Failed to retrieve instance",
			}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (instance *models.Instance, result services.InstanceRetrievalResult) {
					return nil, services.InstanceRetrievalFailure
				},
			}
			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/%s", instanceName), nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToError(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Code).To(Equal(500))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
		})
	})

	_ = Describe("GET instance status", func() {
		_ = It("returns 204 when status is running", func() {
			// arrange
			instanceService := &mocks.InstanceServiceMock{
				GetStatusByNameFunc: func(name string) services.InstanceStatusResult {
					return services.InstanceStatusRunningStatus
				},
			}
			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/%s/status", instanceName), nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(recorder.Code).To(Equal(204))
			Expect(instanceService.GetStatusByNameCalls()).To(HaveLen(1))
		})

		_ = It("returns 404 when instance is not found", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorInstanceStatusRetrievalNotFound,
				Message: "Instance not found",
			}
			instanceService := &mocks.InstanceServiceMock{
				GetStatusByNameFunc: func(name string) services.InstanceStatusResult {
					return services.InstanceStatusNotFound
				},
			}
			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/%s/status", instanceName), nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToError(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Code).To(Equal(404))
			Expect(instanceService.GetStatusByNameCalls()).To(HaveLen(1))
		})

		_ = It("returns 500 when fails to check status", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorInstanceStatusRetrievalFailed,
				Message: "Failed to retrieve instance status",
			}
			instanceService := &mocks.InstanceServiceMock{
				GetStatusByNameFunc: func(name string) services.InstanceStatusResult {
					return services.InstanceStatusFailure
				},
			}
			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/%s/status", instanceName), nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToError(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Code).To(Equal(500))
			Expect(instanceService.GetStatusByNameCalls()).To(HaveLen(1))
		})

		_ = It("returns 500 when instance is in failed status", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorInstanceStatusInstanceFailed,
				Message: "Instance is in failed status",
			}
			instanceService := &mocks.InstanceServiceMock{
				GetStatusByNameFunc: func(name string) services.InstanceStatusResult {
					return services.InstanceStatusFailedStatus
				},
			}
			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/%s/status", instanceName), nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToError(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Code).To(Equal(500))
			Expect(instanceService.GetStatusByNameCalls()).To(HaveLen(1))
		})

		_ = It("returns 202 when instance is in pending status", func() {
			// arrange
			instanceService := &mocks.InstanceServiceMock{
				GetStatusByNameFunc: func(name string) services.InstanceStatusResult {
					return services.InstanceStatusPendingStatus
				},
			}
			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/%s/status", instanceName), nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(recorder.Code).To(Equal(202))
			Expect(instanceService.GetStatusByNameCalls()).To(HaveLen(1))
		})
	})

	_ = Describe("POST instance", func() {
		_ = It("returns 201 when creates successfully", func() {
			// arrange
			instanceService := &mocks.InstanceServiceMock{
				CreateFunc: func(instanceForm *models.InstanceForm) services.InstanceCreationResult {
					return services.InstanceCreationSuccess
				},
			}

			data := url.Values{}
			data.Set("name", instanceForm.Name)
			data.Set("plan", instanceForm.Plan)
			data.Set("team", instanceForm.Team)
			data.Set("user", instanceForm.User)

			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(recorder.Code).To(Equal(201))
			Expect(instanceService.CreateCalls()).To(HaveLen(1))
			Expect(instanceService.CreateCalls()[0].InstanceForm).To(Equal(instanceForm))
		})

		_ = It("returns 409 when instance already exists", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorInstanceCreateAlreadyExists,
				Message: "An instance with this name already exists",
			}

			instanceService := &mocks.InstanceServiceMock{
				CreateFunc: func(instanceForm *models.InstanceForm) services.InstanceCreationResult {
					return services.InstanceCreationAlreadyExist
				},
			}

			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/", nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToError(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Code).To(Equal(409))
			Expect(instanceService.CreateCalls()).To(HaveLen(1))
		})

		_ = It("returns 400 when data is invalid", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorInstanceCreateInvalidData,
				Message: "Invalid instance data",
			}

			instanceService := &mocks.InstanceServiceMock{
				CreateFunc: func(instanceForm *models.InstanceForm) services.InstanceCreationResult {
					return services.InstanceCreationInvalidData
				},
			}

			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/", nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToError(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Code).To(Equal(400))
			Expect(instanceService.CreateCalls()).To(HaveLen(1))
		})

		_ = It("returns 500 when fails to create", func() {
			// arrange
			expected := &models.Error{
				Code:    models.ErrorInstanceCreateFailed,
				Message: "Failed to create instance",
			}

			instanceService := &mocks.InstanceServiceMock{
				CreateFunc: func(instanceForm *models.InstanceForm) services.InstanceCreationResult {
					return services.InstanceCreationFailure
				},
			}

			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/", nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToError(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Code).To(Equal(500))
			Expect(instanceService.CreateCalls()).To(HaveLen(1))
		})

		_ = It("returns 500 when creates successfully but fails to dispatch provision", func() {
			// arrange
			expected := &models.Error{
				Code:    models.ErrorInstanceCreateDispatchProvisionFailed,
				Message: "Instance created, but unable to dispatch provision. Please remove it manually",
			}

			instanceService := &mocks.InstanceServiceMock{
				CreateFunc: func(instanceForm *models.InstanceForm) services.InstanceCreationResult {
					return services.InstanceCreationProvisionFailure
				},
			}

			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/", nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToError(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Code).To(Equal(500))
			Expect(instanceService.CreateCalls()).To(HaveLen(1))
		})
	})

	_ = Describe("DELETE instance", func() {
		_ = It("returns 201 when creates successfully", func() {
			// arrange
			instanceService := &mocks.InstanceServiceMock{
				DeleteFunc: func(name string) services.InstanceDeletionResult {
					return services.InstanceDeletionSuccess
				},
			}

			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%s", instanceName), nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(recorder.Code).To(Equal(200))
			Expect(instanceService.DeleteCalls()).To(HaveLen(1))
		})

		_ = It("returns 404 when instance is not found", func() {
			// arrange
			expected := &models.Error{
				Code:    models.ErrorInstanceDeleteNotFound,
				Message: "Instance not found",
			}

			instanceService := &mocks.InstanceServiceMock{
				DeleteFunc: func(name string) services.InstanceDeletionResult {
					return services.InstanceDeletionNotFound
				},
			}

			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%s", instanceName), nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToError(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Code).To(Equal(404))
			Expect(instanceService.DeleteCalls()).To(HaveLen(1))
		})

		_ = It("returns 500 when instance deletion fails", func() {
			// arrange
			expected := &models.Error{
				Code:    models.ErrorInstanceDeleteFailed,
				Message: "Failed to delete instance",
			}

			instanceService := &mocks.InstanceServiceMock{
				DeleteFunc: func(name string) services.InstanceDeletionResult {
					return services.InstanceDeletionFailure
				},
			}

			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%s", instanceName), nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToError(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Code).To(Equal(500))
			Expect(instanceService.DeleteCalls()).To(HaveLen(1))
		})

		_ = It("returns 500 when instance deletion succeeds but deprovision dispatch fails", func() {
			// arrange
			expected := &models.Error{
				Code:    models.ErrorInstanceDeleteDispatchDeprovisionFailed,
				Message: "Instance deleted, but unable to dispatch deprovision. Please deprovision it manually",
			}

			instanceService := &mocks.InstanceServiceMock{
				DeleteFunc: func(name string) services.InstanceDeletionResult {
					return services.InstanceDeletionDeprovisionFailure
				},
			}

			ginRouter := prepareGinRouter(instanceService, nil)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%s", instanceName), nil)

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			actual := bodyToError(recorder)
			Expect(actual).To(Equal(expected))
			Expect(recorder.Code).To(Equal(500))
			Expect(instanceService.DeleteCalls()).To(HaveLen(1))
		})
	})
})
