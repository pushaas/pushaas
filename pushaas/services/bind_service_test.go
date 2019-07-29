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
	appBindForm := &models.BindAppForm{

	}

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
			varsMap, result := bindService.BindApp(instanceName, appBindForm)

			// assert
			Expect(result).To(Equal(services.AppBindNotFound))
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
			varsMap, result := bindService.BindApp(instanceName, appBindForm)

			// assert
			Expect(result).To(Equal(services.AppBindInstancePending))
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
			varsMap, result := bindService.BindApp(instanceName, appBindForm)

			// assert
			Expect(result).To(Equal(services.AppBindInstanceFailed))
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
						"AppName": "app-1",
						"AppHost": "app-host-1",
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
			varsMap, result := bindService.BindApp(instanceName, appBindForm)

			// assert
			Expect(result).To(Equal(services.AppBindAlreadyBound))
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
			varsMap, result := bindService.BindApp(instanceName, appBindForm)

			// assert
			Expect(result).To(Equal(services.AppBindFailure))
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
			varsMap, result := bindService.BindApp(instanceName, appBindForm)

			// assert
			Expect(result).To(Equal(services.AppBindFailure))
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
			varsMap, result := bindService.BindApp(instanceName, appBindForm)

			// assert
			Expect(result).To(Equal(services.AppBindSuccess))
			Expect(varsMap).To(Equal(expected))
			Expect(instanceService.GetByNameCalls()).To(HaveLen(1))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.HMSetCalls()).To(HaveLen(1))
		})
	})
})
