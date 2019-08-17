package provisioners

import (
	"github.com/rafaeleyng/pushaas/pushaas/models"
)

type (
	PushServiceProvisionResult   int
	PushServiceDeprovisionResult int

	PushServiceProvisioner interface {
		Provision(*models.Instance) PushServiceProvisionResult
		Deprovision(*models.Instance) PushServiceDeprovisionResult
		CleanupServices()
	}
)

const (
	PushServiceProvisionResultSuccess PushServiceProvisionResult = iota
	PushServiceProvisionResultFailure
)

const (
	PushServiceDeprovisionResultSuccess PushServiceDeprovisionResult = iota
	PushServiceDeprovisionResultFailure
)
