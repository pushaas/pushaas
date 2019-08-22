package services_test

import (
	"errors"

	"github.com/go-redis/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/pushaas/pushaas/pushaas/mocks"
	"github.com/pushaas/pushaas/pushaas/models"
	"github.com/pushaas/pushaas/pushaas/services"
)

var _ = Describe("BindService", func() {
	config := viper.New()
	instanceName := "instance-1"
	bindAppForm := &models.BindAppForm{}
	bindUnitForm := &models.BindUnitForm{}
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

	_ = Describe("BindUnit", func() {
		_ = It("indicates when fails to check existing app bind", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, errors.New("some error"))
				},
			}
			instanceService := &mocks.InstanceServiceMock{}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.BindUnit(instanceName, bindUnitForm)

			// assert
			Expect(result).To(Equal(services.BindUnitFailure))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.SAddCalls()).To(HaveLen(0))
		})

		_ = It("indicates when app bind is not found", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, nil)
				},
			}
			instanceService := &mocks.InstanceServiceMock{}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.BindUnit(instanceName, bindUnitForm)

			// assert
			Expect(result).To(Equal(services.BindUnitAppNotBound))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.SAddCalls()).To(HaveLen(0))
		})

		_ = It("indicates when fails to bind unit", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(map[string]string{
						"AppName": appName,
						"AppHost": appHost,
					}, nil)
				},
				SAddFunc: func(key string, members ...interface{}) *redis.IntCmd {
					return redis.NewIntResult(0, errors.New("some error"))
				},
			}
			instanceService := &mocks.InstanceServiceMock{}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.BindUnit(instanceName, bindUnitForm)

			// assert
			Expect(result).To(Equal(services.BindUnitFailure))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.SAddCalls()).To(HaveLen(1))
		})

		_ = It("indicates when unit is already bound", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(map[string]string{
						"AppName": appName,
						"AppHost": appHost,
					}, nil)
				},
				SAddFunc: func(key string, members ...interface{}) *redis.IntCmd {
					return redis.NewIntResult(0, nil)
				},
			}
			instanceService := &mocks.InstanceServiceMock{}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.BindUnit(instanceName, bindUnitForm)

			// assert
			Expect(result).To(Equal(services.BindUnitAlreadyBound))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.SAddCalls()).To(HaveLen(1))
		})

		_ = It("indicates when binds unit successfully", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(map[string]string{
						"AppName": appName,
						"AppHost": appHost,
					}, nil)
				},
				SAddFunc: func(key string, members ...interface{}) *redis.IntCmd {
					return redis.NewIntResult(1, nil)
				},
			}
			instanceService := &mocks.InstanceServiceMock{}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.BindUnit(instanceName, bindUnitForm)

			// assert
			Expect(result).To(Equal(services.BindUnitSuccess))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.SAddCalls()).To(HaveLen(1))
		})
	})

	_ = Describe("UnbindUnit", func() {
		_ = It("indicates when fails to check existing app bind", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, errors.New("some error"))
				},
			}
			instanceService := &mocks.InstanceServiceMock{}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.UnbindUnit(instanceName, bindUnitForm)

			// assert
			Expect(result).To(Equal(services.UnbindUnitFailure))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.SRemCalls()).To(HaveLen(0))
		})

		_ = It("indicates when app bind is not found", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, nil)
				},
			}
			instanceService := &mocks.InstanceServiceMock{}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.UnbindUnit(instanceName, bindUnitForm)

			// assert
			Expect(result).To(Equal(services.UnbindUnitAppNotBound))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.SRemCalls()).To(HaveLen(0))
		})

		_ = It("indicates when fails to unbind unit", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(map[string]string{
						"AppName": appName,
						"AppHost": appHost,
					}, nil)
				},
				SRemFunc: func(key string, members ...interface{}) *redis.IntCmd {
					return redis.NewIntResult(0, errors.New("some error"))
				},
			}
			instanceService := &mocks.InstanceServiceMock{}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.UnbindUnit(instanceName, bindUnitForm)

			// assert
			Expect(result).To(Equal(services.UnbindUnitFailure))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.SRemCalls()).To(HaveLen(1))
		})

		_ = It("indicates when unit is not bound", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(map[string]string{
						"AppName": appName,
						"AppHost": appHost,
					}, nil)
				},
				SRemFunc: func(key string, members ...interface{}) *redis.IntCmd {
					return redis.NewIntResult(0, nil)
				},
			}
			instanceService := &mocks.InstanceServiceMock{}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.UnbindUnit(instanceName, bindUnitForm)

			// assert
			Expect(result).To(Equal(services.UnbindUnitNotBound))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.SRemCalls()).To(HaveLen(1))
		})

		_ = It("indicates when unbinds unit successfully", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(map[string]string{
						"AppName": appName,
						"AppHost": appHost,
					}, nil)
				},
				SRemFunc: func(key string, members ...interface{}) *redis.IntCmd {
					return redis.NewIntResult(1, nil)
				},
			}
			instanceService := &mocks.InstanceServiceMock{}
			bindService := services.NewBindService(config, logger, redisClient, instanceService)

			// act
			result := bindService.UnbindUnit(instanceName, bindUnitForm)

			// assert
			Expect(result).To(Equal(services.UnbindUnitSuccess))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.SRemCalls()).To(HaveLen(1))
		})
	})
})
