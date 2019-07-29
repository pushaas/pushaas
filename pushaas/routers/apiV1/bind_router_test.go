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


var _ = Describe("BindRouter", func() {
	instanceName := "instance-1"
	bindAppForm := &models.BindAppForm{
		AppName: "app-1",
		AppHost: "app-host-1",
	}

	prepareGinRouter := func(bindService services.BindService) *gin.Engine {
		ginRouter := gin.New()
		router := apiV1.NewBindRouter(bindService)
		router.SetupRoutes(ginRouter)
		return ginRouter
	}

	bodyToError := func(recorder *httptest.ResponseRecorder) *models.Error {
		var body *models.Error
		_ = json.Unmarshal([]byte(recorder.Body.String()), &body)
		return body
	}

	bodyToVarsMap := func(recorder *httptest.ResponseRecorder) map[string]string {
		var body map[string]string
		_ = json.Unmarshal([]byte(recorder.Body.String()), &body)
		return body
	}

	_ = Describe("POST bindApp", func() {
		_ = It("returns 201 when creates successfully", func() {
			// arrange
			expected := map[string]string {
				"env1": "value1",
				"env2": "value2",
			}
			bindService := &mocks.BindServiceMock{
				BindAppFunc: func(name string, bindAppForm *models.BindAppForm) (map[string]string, services.BindAppResult) {
					return expected, services.BindAppSuccess
				},
			}

			data := url.Values{}
			data.Set("app-name", bindAppForm.AppName)
			data.Set("app-host", bindAppForm.AppHost)

			ginRouter := prepareGinRouter(bindService)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", fmt.Sprintf("/%s/bind-app", instanceName), strings.NewReader(data.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(bodyToVarsMap(recorder)).To(Equal(expected))
			Expect(recorder.Code).To(Equal(201))
			Expect(bindService.BindAppCalls()).To(HaveLen(1))
		})

		_ = It("returns 404 when instance is not found", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorBindAppNotFound,
				Message: "Instance not found",
			}

			bindService := &mocks.BindServiceMock{
				BindAppFunc: func(name string, bindAppForm *models.BindAppForm) (map[string]string, services.BindAppResult) {
					return nil, services.BindAppNotFound
				},
			}

			data := url.Values{}
			data.Set("app-name", bindAppForm.AppName)
			data.Set("app-host", bindAppForm.AppHost)

			ginRouter := prepareGinRouter(bindService)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", fmt.Sprintf("/%s/bind-app", instanceName), strings.NewReader(data.Encode()))

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(bodyToError(recorder)).To(Equal(expected))
			Expect(recorder.Code).To(Equal(404))
			Expect(bindService.BindAppCalls()).To(HaveLen(1))
		})

		_ = It("returns 412 when instance is in pending status", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorBindAppInstancePending,
				Message: "Instance is in pending status",
			}

			bindService := &mocks.BindServiceMock{
				BindAppFunc: func(name string, bindAppForm *models.BindAppForm) (map[string]string, services.BindAppResult) {
					return nil, services.BindAppInstancePending
				},
			}

			data := url.Values{}
			data.Set("app-name", bindAppForm.AppName)
			data.Set("app-host", bindAppForm.AppHost)

			ginRouter := prepareGinRouter(bindService)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", fmt.Sprintf("/%s/bind-app", instanceName), strings.NewReader(data.Encode()))

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(bodyToError(recorder)).To(Equal(expected))
			Expect(recorder.Code).To(Equal(412))
			Expect(bindService.BindAppCalls()).To(HaveLen(1))
		})

		_ = It("returns 500 when instance is in failed status", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorBindAppInstanceFailed,
				Message: "Instance is in failed status",
			}

			bindService := &mocks.BindServiceMock{
				BindAppFunc: func(name string, bindAppForm *models.BindAppForm) (map[string]string, services.BindAppResult) {
					return nil, services.BindAppInstanceFailed
				},
			}

			data := url.Values{}
			data.Set("app-name", bindAppForm.AppName)
			data.Set("app-host", bindAppForm.AppHost)

			ginRouter := prepareGinRouter(bindService)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", fmt.Sprintf("/%s/bind-app", instanceName), strings.NewReader(data.Encode()))

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(bodyToError(recorder)).To(Equal(expected))
			Expect(recorder.Code).To(Equal(500))
			Expect(bindService.BindAppCalls()).To(HaveLen(1))
		})

		_ = It("returns 500 when instance is already bound to app", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorBindAppAlreadyBound,
				Message: "Instance is already bound to app",
			}

			bindService := &mocks.BindServiceMock{
				BindAppFunc: func(name string, bindAppForm *models.BindAppForm) (map[string]string, services.BindAppResult) {
					return nil, services.BindAppAlreadyBound
				},
			}

			data := url.Values{}
			data.Set("app-name", bindAppForm.AppName)
			data.Set("app-host", bindAppForm.AppHost)

			ginRouter := prepareGinRouter(bindService)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", fmt.Sprintf("/%s/bind-app", instanceName), strings.NewReader(data.Encode()))

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(bodyToError(recorder)).To(Equal(expected))
			Expect(recorder.Code).To(Equal(500))
			Expect(bindService.BindAppCalls()).To(HaveLen(1))
		})

		_ = It("returns 500 when fails to bind instance to app", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorBindAppFailed,
				Message: "Failed to bind instance to app",
			}

			bindService := &mocks.BindServiceMock{
				BindAppFunc: func(name string, bindAppForm *models.BindAppForm) (map[string]string, services.BindAppResult) {
					return nil, services.BindAppFailure
				},
			}

			data := url.Values{}
			data.Set("app-name", bindAppForm.AppName)
			data.Set("app-host", bindAppForm.AppHost)

			ginRouter := prepareGinRouter(bindService)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", fmt.Sprintf("/%s/bind-app", instanceName), strings.NewReader(data.Encode()))

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(bodyToError(recorder)).To(Equal(expected))
			Expect(recorder.Code).To(Equal(500))
			Expect(bindService.BindAppCalls()).To(HaveLen(1))
		})
	})

	_ = Describe("DELETE instance", func() {
		_ = It("returns 200 when unbinds successfully", func() {
			// arrange
			bindService := &mocks.BindServiceMock{
				UnbindAppFunc: func(name string, bindAppForm *models.BindAppForm) services.AppUnbindResult {
					return services.AppUnbindSuccess
				},
			}

			data := url.Values{}
			data.Set("app-name", bindAppForm.AppName)
			data.Set("app-host", bindAppForm.AppHost)

			ginRouter := prepareGinRouter(bindService)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%s/bind-app", instanceName), strings.NewReader(data.Encode()))

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(recorder.Code).To(Equal(200))
			Expect(bindService.UnbindAppCalls()).To(HaveLen(1))
		})

		_ = It("returns 404 when instance is not found", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorUnbindAppNotFound,
				Message: "Instance not found",
			}

			bindService := &mocks.BindServiceMock{
				UnbindAppFunc: func(name string, bindAppForm *models.BindAppForm) services.AppUnbindResult {
					return services.AppUnbindInstanceNotFound
				},
			}

			data := url.Values{}
			data.Set("app-name", bindAppForm.AppName)
			data.Set("app-host", bindAppForm.AppHost)

			ginRouter := prepareGinRouter(bindService)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%s/bind-app", instanceName), strings.NewReader(data.Encode()))

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(bodyToError(recorder)).To(Equal(expected))
			Expect(recorder.Code).To(Equal(404))
			Expect(bindService.UnbindAppCalls()).To(HaveLen(1))
		})

		_ = It("returns 404 when binding is not found", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorUnbindAppNotBound,
				Message: "Instance is not bound to app",
			}

			bindService := &mocks.BindServiceMock{
				UnbindAppFunc: func(name string, bindAppForm *models.BindAppForm) services.AppUnbindResult {
					return services.AppUnbindNotBound
				},
			}

			data := url.Values{}
			data.Set("app-name", bindAppForm.AppName)
			data.Set("app-host", bindAppForm.AppHost)

			ginRouter := prepareGinRouter(bindService)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%s/bind-app", instanceName), strings.NewReader(data.Encode()))

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(bodyToError(recorder)).To(Equal(expected))
			Expect(recorder.Code).To(Equal(404))
			Expect(bindService.UnbindAppCalls()).To(HaveLen(1))
		})

		_ = It("returns 500 when fails to unbind", func() {
			// arrange
			expected := &models.Error{
				Code: models.ErrorUnbindAppFailed,
				Message: "Failed to unbind instance from app",
			}

			bindService := &mocks.BindServiceMock{
				UnbindAppFunc: func(name string, bindAppForm *models.BindAppForm) services.AppUnbindResult {
					return services.AppUnbindFailure
				},
			}

			data := url.Values{}
			data.Set("app-name", bindAppForm.AppName)
			data.Set("app-host", bindAppForm.AppHost)

			ginRouter := prepareGinRouter(bindService)
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%s/bind-app", instanceName), strings.NewReader(data.Encode()))

			// act
			ginRouter.ServeHTTP(recorder, req)

			// assert
			Expect(bodyToError(recorder)).To(Equal(expected))
			Expect(recorder.Code).To(Equal(500))
			Expect(bindService.UnbindAppCalls()).To(HaveLen(1))
		})
	})
})
