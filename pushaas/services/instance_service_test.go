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

var _ = Describe("InstanceService", func() {
	config := viper.New()
	instanceName := "instance-1"
	instanceForm := &models.InstanceForm{
		Name: instanceName,
		Team: "pushaas-team",
		User: "rafael",
		Plan: string(models.PlanSmall),
	}

	Describe("GetByName", func() {
		It("should return instance and success code when no errors occur", func() {
			// arrange
			expected := &models.Instance{
				Name: instanceName,
				Team: "pushaas-team",
				User: "rafael",
			}
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					val := map[string]string{
						"Name": expected.Name,
						"Team": expected.Team,
						"User": expected.User,
					}
					return redis.NewStringStringMapResult(val, nil)
				},
			}

			instanceService := services.NewInstanceService(config, logger, redisClient, nil)

			// act
			instance, result := instanceService.GetByName(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceRetrievalSuccess))
			Expect(instance).To(Equal(expected))
		})

		It("indicates when instance is not found", func() {
			// arrange
			var expected *models.Instance
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, nil)
				},
			}
			instanceService := services.NewInstanceService(config, logger, redisClient, nil)

			// act
			instance, result := instanceService.GetByName(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceRetrievalNotFound))
			Expect(instance).To(Equal(expected))
		})

		It("indicates when failure happens in retrieving the instance", func() {
			// arrange
			var expected *models.Instance
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, errors.New("some error"))
				},
			}
			instanceService := services.NewInstanceService(config, logger, redisClient, nil)

			// act
			instance, result := instanceService.GetByName(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceRetrievalFailure))
			Expect(instance).To(Equal(expected))
		})
	})

	Describe("GetStatusByName", func() {
		It("indicates when instance is not found", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, nil)
				},
			}
			instanceService := services.NewInstanceService(config, logger, redisClient, nil)

			// act
			result := instanceService.GetStatusByName(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceStatusNotFound))
		})

		It("indicates when fails to get instance", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, errors.New("some error"))
				},
			}
			instanceService := services.NewInstanceService(config, logger, redisClient, nil)

			// act
			result := instanceService.GetStatusByName(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceStatusFailure))
		})

		It("indicates when gets instance and is on status pending", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					val := map[string]string{
						"Status": string(models.InstanceStatusPending),
					}
					return redis.NewStringStringMapResult(val, nil)
				},			}
			instanceService := services.NewInstanceService(config, logger, redisClient, nil)

			// act
			result := instanceService.GetStatusByName(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceStatusPendingStatus))
		})

		It("indicates when gets instance and is on status failed", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					val := map[string]string{
						"Status": string(models.InstanceStatusFailed),
					}
					return redis.NewStringStringMapResult(val, nil)
				},			}
			instanceService := services.NewInstanceService(config, logger, redisClient, nil)

			// act
			result := instanceService.GetStatusByName(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceStatusFailedStatus))
		})

		It("indicates when gets instance and is on status running", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					val := map[string]string{
						"Status": string(models.InstanceStatusRunning),
					}
					return redis.NewStringStringMapResult(val, nil)
				},
			}
			instanceService := services.NewInstanceService(config, logger, redisClient, nil)

			// act
			result := instanceService.GetStatusByName(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceStatusRunningStatus))
		})
	})

	Describe("Delete", func() {
		It("indicates when instance is not found at retrieval", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, nil)
				},
			}
			provisionService := &mocks.ProvisionServiceMock{}
			instanceService := services.NewInstanceService(config, logger, redisClient, provisionService)

			// act
			result := instanceService.Delete(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceDeletionNotFound))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.DelCalls()).To(HaveLen(0))
			Expect(provisionService.DispatchDeprovisionCalls()).To(HaveLen(0))
		})

		It("indicates when failed to get instance to delete", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, errors.New("some error"))
				},
			}
			provisionService := &mocks.ProvisionServiceMock{}
			instanceService := services.NewInstanceService(config, logger, redisClient, provisionService)

			// act
			result := instanceService.Delete(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceDeletionFailure))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.DelCalls()).To(HaveLen(0))
			Expect(provisionService.DispatchDeprovisionCalls()).To(HaveLen(0))
		})

		It("indicates when failed to delete instance", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					val := map[string]string {
						"Status": string(models.InstanceStatusRunning),
					}
					return redis.NewStringStringMapResult(val, nil)
				},
				DelFunc: func(keys ...string) *redis.IntCmd {
					return redis.NewIntResult(0, errors.New("some error"))
				},
			}
			provisionService := &mocks.ProvisionServiceMock{}
			instanceService := services.NewInstanceService(config, logger, redisClient, provisionService)

			// act
			result := instanceService.Delete(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceDeletionFailure))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.DelCalls()).To(HaveLen(1))
			Expect(provisionService.DispatchDeprovisionCalls()).To(HaveLen(0))
		})

		It("indicates when instance is not found at deletion", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					val := map[string]string {
						"Status": string(models.InstanceStatusRunning),
					}
					return redis.NewStringStringMapResult(val, nil)
				},
				DelFunc: func(keys ...string) *redis.IntCmd {
					return redis.NewIntResult(0, nil)
				},
			}
			provisionService := &mocks.ProvisionServiceMock{}
			instanceService := services.NewInstanceService(config, logger, redisClient, provisionService)

			// act
			result := instanceService.Delete(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceDeletionNotFound))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.DelCalls()).To(HaveLen(1))
			Expect(provisionService.DispatchDeprovisionCalls()).To(HaveLen(0))
		})

		It("indicates when deletes successfully but fails to dispatch deprovision", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					val := map[string]string {
						"Status": string(models.InstanceStatusRunning),
					}
					return redis.NewStringStringMapResult(val, nil)
				},
				DelFunc: func(keys ...string) *redis.IntCmd {
					return redis.NewIntResult(1, nil)
				},
			}
			provisionService := &mocks.ProvisionServiceMock{
				DispatchDeprovisionFunc: func(instance *models.Instance) services.DispatchDeprovisionResult {
					return services.DispatchDeprovisionResultFailure
				},
			}
			instanceService := services.NewInstanceService(config, logger, redisClient, provisionService)

			// act
			result := instanceService.Delete(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceDeletionDeprovisionFailure))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.DelCalls()).To(HaveLen(1))
			Expect(provisionService.DispatchDeprovisionCalls()).To(HaveLen(1))
		})

		It("indicates when deletes successfully and dispatches deprovision", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					val := map[string]string {
						"Status": string(models.InstanceStatusRunning),
					}
					return redis.NewStringStringMapResult(val, nil)
				},
				DelFunc: func(keys ...string) *redis.IntCmd {
					return redis.NewIntResult(1, nil)
				},
			}
			provisionService := &mocks.ProvisionServiceMock{
				DispatchDeprovisionFunc: func(instance *models.Instance) services.DispatchDeprovisionResult {
					return services.DispatchDeprovisionResultSuccess
				},
			}
			instanceService := services.NewInstanceService(config, logger, redisClient, provisionService)

			// act
			result := instanceService.Delete(instanceName)

			// assert
			Expect(result).To(Equal(services.InstanceDeletionSuccess))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.DelCalls()).To(HaveLen(1))
			Expect(provisionService.DispatchDeprovisionCalls()).To(HaveLen(1))
		})
	})

	Describe("Create", func() {
		It("indicates when instance with same name already exists", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					val := map[string]string {
						"Status": string(models.InstanceStatusRunning),
					}
					return redis.NewStringStringMapResult(val, nil)
				},
			}
			provisionService := &mocks.ProvisionServiceMock{}
			instanceService := services.NewInstanceService(config, logger, redisClient, provisionService)

			// act
			result := instanceService.Create(instanceForm)

			// assert
			Expect(result).To(Equal(services.InstanceCreationAlreadyExist))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.HSetCalls()).To(HaveLen(0))
			Expect(provisionService.DispatchProvisionCalls()).To(HaveLen(0))
		})

		It("indicates when fails to check instance existence", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, errors.New("some error"))
				},
			}
			provisionService := &mocks.ProvisionServiceMock{}
			instanceService := services.NewInstanceService(config, logger, redisClient, provisionService)

			// act
			result := instanceService.Create(instanceForm)

			// assert
			Expect(result).To(Equal(services.InstanceCreationFailure))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.HSetCalls()).To(HaveLen(0))
			Expect(provisionService.DispatchProvisionCalls()).To(HaveLen(0))
		})

		It("indicates when fails to check instance existence", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, nil)
				},
			}
			provisionService := &mocks.ProvisionServiceMock{}
			instanceService := services.NewInstanceService(config, logger, redisClient, provisionService)
			instanceFormInvalid := &models.InstanceForm{}

			// act
			result := instanceService.Create(instanceFormInvalid)

			// assert
			Expect(result).To(Equal(services.InstanceCreationInvalidData))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.HSetCalls()).To(HaveLen(0))
			Expect(provisionService.DispatchProvisionCalls()).To(HaveLen(0))
		})

		It("indicates when fails to create instance", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, nil)
				},
				HMSetFunc: func(key string, fields map[string]interface{}) *redis.StatusCmd {
					return redis.NewStatusResult("", errors.New("some error"))
				},
			}
			provisionService := &mocks.ProvisionServiceMock{}
			instanceService := services.NewInstanceService(config, logger, redisClient, provisionService)

			// act
			result := instanceService.Create(instanceForm)

			// assert
			Expect(result).To(Equal(services.InstanceCreationFailure))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.HMSetCalls()).To(HaveLen(1))
			Expect(provisionService.DispatchProvisionCalls()).To(HaveLen(0))
		})

		It("indicates when creates instance but fails to dispatch provision", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, nil)
				},
				HMSetFunc: func(key string, fields map[string]interface{}) *redis.StatusCmd {
					return redis.NewStatusResult("", nil)
				},
			}
			provisionService := &mocks.ProvisionServiceMock{
				DispatchProvisionFunc: func(instance *models.Instance) services.DispatchProvisionResult {
					return services.DispatchProvisionResultFailure
				},
			}
			instanceService := services.NewInstanceService(config, logger, redisClient, provisionService)

			// act
			result := instanceService.Create(instanceForm)

			// assert
			Expect(result).To(Equal(services.InstanceCreationProvisionFailure))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.HMSetCalls()).To(HaveLen(1))
			Expect(provisionService.DispatchProvisionCalls()).To(HaveLen(1))
		})

		It("indicates when creates instance and dispatches provision successfully", func() {
			// arrange
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					return redis.NewStringStringMapResult(nil, nil)
				},
				HMSetFunc: func(key string, fields map[string]interface{}) *redis.StatusCmd {
					return redis.NewStatusResult("", nil)
				},
			}
			provisionService := &mocks.ProvisionServiceMock{
				DispatchProvisionFunc: func(instance *models.Instance) services.DispatchProvisionResult {
					return services.DispatchProvisionResultSuccess
				},
			}
			instanceService := services.NewInstanceService(config, logger, redisClient, provisionService)

			// act
			result := instanceService.Create(instanceForm)

			// assert
			Expect(result).To(Equal(services.InstanceCreationSuccess))
			Expect(redisClient.HGetAllCalls()).To(HaveLen(1))
			Expect(redisClient.HMSetCalls()).To(HaveLen(1))
			Expect(provisionService.DispatchProvisionCalls()).To(HaveLen(1))
		})
	})
})
