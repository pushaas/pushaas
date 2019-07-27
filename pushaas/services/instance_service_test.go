package services_test

import (
	"github.com/go-redis/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/mocks"
	"github.com/rafaeleyng/pushaas/pushaas/models"
	"github.com/rafaeleyng/pushaas/pushaas/services"
)

var _ = Describe("InstanceService", func() {
	Describe("GetByName", func() {
		It("should get instance correctly", func() {
			// arrange
			expected := &models.Instance{
				Name: "instance-1",
				Team: "pushaas-team",
				User: "rafael",
			}
			config := viper.New()
			var logger *zap.Logger
			redisClient := &mocks.UniversalClientMock{
				HGetAllFunc: func(key string) *redis.StringStringMapCmd {
					val := map[string]string{
						"Name": expected.Name,
						"Team": expected.Team,
						"User": expected.User,
					}
					var err error = nil
					return redis.NewStringStringMapResult(val, err)
				},
			}
			var provisionService *mocks.ProvisionServiceMock

			instanceService := services.NewInstanceService(config, logger, redisClient, provisionService)

			// act
			instance, result := instanceService.GetByName("instance-1")

			// assert
			Expect(result).To(Equal(services.InstanceRetrievalSuccess))
			Expect(instance).To(Equal(expected))
		})
	})
})
