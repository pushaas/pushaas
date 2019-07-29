package provisioners

import (
	"github.com/rafaeleyng/pushaas/pushaas/models"
)

type (
	ProvisionResult   int
	DeprovisionResult int

	Provisioner interface {
		Provision(*models.Instance) ProvisionResult
		Deprovision(*models.Instance) DeprovisionResult
	}
)

const (
	ProvisionResultSuccess ProvisionResult = iota
)

const (
	DeprovisionResultSuccess DeprovisionResult = iota
)
