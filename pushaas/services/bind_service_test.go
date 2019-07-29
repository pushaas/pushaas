package services_test

import (
	"errors"

	"github.com/go-redis/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/rafaeleyng/pushaas/pushaas/mocks"
	"github.com/rafaeleyng/pushaas/pushaas/models"
	"github.com/rafaeleyng/pushaas/pushaas/services"
)

var _ = Describe("BindService", func() {
	config := viper.New()
	instanceName := "instance-1"
	bindAppForm := &models.BindAppForm{}
	appName := "app-1"
	appHost := "app-host-1"

	_ = Describe("BindApp", func() {
		_ = It("indicates when instance is not found", func() {
			// arrange
			var expected map[string]string
			redisClient := &mocks.UniversalClientMock{}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (*models.Instance, services.InstanceRetrievalResult) {
					return nil, services.InstanceRetrievalNotFound
				},
			}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			varsMap, result := bindService.BindApp(instanceName, bindAppForm)

			// assert
			Expect(result).To(Equal(services.BindAppNotFound))
			Expect(varsMap).To(Equal(expected))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
		})

		_ = It("indicates when instance is in pending status", func() {
			// arrange
			var expected map[string]string
			instance := &models.Instance{
				Status: models.InstanceStatusPending,
			}
			redisClient := &mocks.UniversalClientMock{}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (*models.Instance, services.InstanceRetrievalResult) {
					return instance, services.InstanceRetrievalSuccess
				},
			}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			varsMap, result := bindService.BindApp(instanceName, bindAppForm)

			// assert
			Expect(result).To(Equal(services.BindAppInstancePending))
			Expect(varsMap).To(Equal(expected))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
		})

		_ = It("indicates when instance is in pending status", func() {
			// arrange
			var expected map[string]string
			instance := &models.Instance{
				Status: models.InstanceStatusFailed,
			}
			redisClient := &mocks.UniversalClientMock{}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (*models.Instance, services.InstanceRetrievalResult) {
					return instance, services.InstanceRetrievalSuccess
				},
			}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			varsMap, result := bindService.BindApp(instanceName, bindAppForm)

			// assert
			Expect(result).To(Equal(services.BindAppInstanceFailed))
			Expect(varsMap).To(Equal(expected))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(0))
			Expect(redisClient.HMSetCalls()).To(HaveLen(0))
		})

		_ = It("indicates when instance is already bound to an app", func() {
			// arrange
			var expected map[string]string
			instance := &models.Instance{
				Status: models.InstanceStatusRunning,
			}
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(map[string]string {
						"AppName": appName,
						"AppHost": appHost,
					}, nil)
				},
			}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (*models.Instance, services.InstanceRetrievalResult) {
					return instance, services.InstanceRetrievalSuccess
				},
			}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			varsMap, result := bindService.BindApp(instanceName, bindAppForm)

			// assert
			Expect(result).To(Equal(services.BindAppAlreadyBound))
			Expect(varsMap).To(Equal(expected))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.HMSetCalls()).To(HaveLen(0))
		})

		_ = It("indicates when fails to check existing bind", func() {
			// arrange
			var expected map[string]string
			instance := &models.Instance{
				Status: models.InstanceStatusRunning,
			}
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, errors.New("some error"))
				},
			}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (*models.Instance, services.InstanceRetrievalResult) {
					return instance, services.InstanceRetrievalSuccess
				},
			}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			varsMap, result := bindService.BindApp(instanceName, bindAppForm)

			// assert
			Expect(result).To(Equal(services.BindAppFailure))
			Expect(varsMap).To(Equal(expected))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.HMSetCalls()).To(HaveLen(0))
		})

		_ = It("indicates when fails to create new bind", func() {
			// arrange
			var expected map[string]string
			instance := &models.Instance{
				Status: models.InstanceStatusRunning,
			}
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(map[string]string {}, nil)
				},
				HMSetFunc: func(key string, fields map[string]interface{}) *redis.StatusCmd {
					return redis.NewStatusResult("", errors.New("some error"))
				},
			}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (*models.Instance, services.InstanceRetrievalResult) {
					return instance, services.InstanceRetrievalSuccess
				},
			}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			varsMap, result := bindService.BindApp(instanceName, bindAppForm)

			// assert
			Expect(result).To(Equal(services.BindAppFailure))
			Expect(varsMap).To(Equal(expected))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.HMSetCalls()).To(HaveLen(1))
		})

		_ = It("indicates when creates new bind successfully", func() {
			// arrange
			expected := map[string]string{
				"PUSHAAS_ENDPOINT": "the-endpoint",
				"PUSHAAS_USERNAME": "the-username",
				"PUSHAAS_PASSWORD": "the-password",
			}
			instance := &models.Instance{
				Status: models.InstanceStatusRunning,
			}
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(map[string]string {}, nil)
				},
				HMSetFunc: func(key string, fields map[string]interface{}) *redis.StatusCmd {
					return redis.NewStatusResult("ok", nil)
				},
			}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (*models.Instance, services.InstanceRetrievalResult) {
					return instance, services.InstanceRetrievalSuccess
				},
			}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			varsMap, result := bindService.BindApp(instanceName, bindAppForm)

			// assert
			Expect(result).To(Equal(services.BindAppSuccess))
			Expect(varsMap).To(Equal(expected))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.HMSetCalls()).To(HaveLen(1))
		})
	})

	_ = Describe("UnbindApp", func() {
		_ = It("indicates when instance does not exist", func() {
			// arrange
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (*models.Instance, services.InstanceRetrievalResult) {
					return nil, services.InstanceRetrievalNotFound
				},
			}
			redisClient := &mocks.UniversalClientMock{}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.UnbindApp(instanceName, bindAppForm)

			// assert
			Expect(result).To(Equal(services.UnbindAppInstanceNotFound))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(0))
			Expect(redisClient.DelCalls()).To(HaveLen(0))
		})

		_ = It("indicates when fails to check existing app binding", func() {
			// arrange
			instance := &models.Instance{
				Name: instanceName,
				Status: models.InstanceStatusRunning,
			}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (*models.Instance, services.InstanceRetrievalResult) {
					return instance, services.InstanceRetrievalSuccess
				},
			}
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, errors.New("some error"))
				},
			}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.UnbindApp(instanceName, bindAppForm)

			// assert
			Expect(result).To(Equal(services.UnbindAppFailure))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.DelCalls()).To(HaveLen(0))
		})

		_ = It("indicates when can't find instance binding to app", func() {
			// arrange
			instance := &models.Instance{
				Name: instanceName,
				Status: models.InstanceStatusRunning,
			}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (*models.Instance, services.InstanceRetrievalResult) {
					return instance, services.InstanceRetrievalSuccess
				},
			}
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(map[string]string{}, nil)
				},
			}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.UnbindApp(instanceName, bindAppForm)

			// assert
			Expect(result).To(Equal(services.UnbindAppNotBound))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.DelCalls()).To(HaveLen(0))
		})

		_ = It("indicates when fails to remove binding", func() {
			// arrange
			instance := &models.Instance{
				Name: instanceName,
				Status: models.InstanceStatusRunning,
			}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (*models.Instance, services.InstanceRetrievalResult) {
					return instance, services.InstanceRetrievalSuccess
				},
			}
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(map[string]string{
						"AppName": appName,
						"AppHost": appHost,
					}, nil)
				},
				DelFunc: func(keys ...string) *redis.IntCmd {
					return redis.NewIntResult(0, errors.New("some error"))
				},
			}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.UnbindApp(instanceName, bindAppForm)

			// assert
			Expect(result).To(Equal(services.UnbindAppFailure))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.DelCalls()).To(HaveLen(1))
		})

		_ = It("indicates when does not find the binding when tries to remove it", func() {
			// arrange
			instance := &models.Instance{
				Name: instanceName,
				Status: models.InstanceStatusRunning,
			}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (*models.Instance, services.InstanceRetrievalResult) {
					return instance, services.InstanceRetrievalSuccess
				},
			}
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(map[string]string{
						"AppName": appName,
						"AppHost": appHost,
					}, nil)
				},
				DelFunc: func(keys ...string) *redis.IntCmd {
					return redis.NewIntResult(0, nil)
				},
			}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.UnbindApp(instanceName, bindAppForm)

			// assert
			Expect(result).To(Equal(services.UnbindAppNotBound))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.DelCalls()).To(HaveLen(1))
		})

		_ = It("indicates when removes the binding successfully", func() {
			// arrange
			instance := &models.Instance{
				Name: instanceName,
				Status: models.InstanceStatusRunning,
			}
			instanceService := &mocks.InstanceServiceMock{
				GetByNameFunc: func(name string) (*models.Instance, services.InstanceRetrievalResult) {
					return instance, services.InstanceRetrievalSuccess
				},
			}
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(map[string]string{
						"AppName": appName,
						"AppHost": appHost,
					}, nil)
				},
				DelFunc: func(keys ...string) *redis.IntCmd {
					return redis.NewIntResult(1, nil)
				},
			}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.UnbindApp(instanceName, bindAppForm)

			// assert
			Expect(result).To(Equal(services.UnbindAppSuccess))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.DelCalls()).To(HaveLen(1))
		})
	})
})
