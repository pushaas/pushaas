package services

import (
	"github.com/rafaeleyng/pushaas/pushaas/models"
)

type (
	PlanService interface {
		GetAll() []models.Plan
	}

	planService struct{}
)

func (s *planService) GetAll() []models.Plan {
	result := []models.Plan{
		{
			Name:        models.PlanSmall,
			Description: "The only plan",
		},
	}

	return result
}
func NewPlanService() PlanService {
	return &planService{}
}
