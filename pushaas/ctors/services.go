package ctors

import (
	"github.com/go-bongo/bongo"

	"github.com/rafaeleyng/pushaas/pushaas/services"
)

func NewInstanceService(mongodb *bongo.Connection) services.InstanceService {
	return services.NewInstanceService(mongodb)
}
