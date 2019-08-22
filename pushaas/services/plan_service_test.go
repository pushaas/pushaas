package services_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pushaas/pushaas/pushaas/services"
)

var _ = Describe("PlanService", func() {
	Describe("GetAll", func() {
		It("should return the available plans", func() {
			// arrange
			planService := services.NewPlanService()

			// act
			plans := planService.GetAll()

			// assert
			Expect(len(plans)).To(Equal(1))
			Expect(plans[0].Name).To(Equal("small"))
			Expect(plans[0].Description).To(Equal("The only plan"))
		})
	})
})
