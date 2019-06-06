package ctors

import (
	"github.com/go-bongo/bongo"
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/provisioners"
	"github.com/rafaeleyng/pushaas/pushaas/services"
)

func NewInstanceService(logger *zap.Logger, mongodb *bongo.Connection, provisioner provisioners.Provisioner) services.InstanceService {
	return services.NewInstanceService(logger, mongodb, provisioner)
}

func NewPlanService() services.PlanService {
	return services.NewPlanService()
}
