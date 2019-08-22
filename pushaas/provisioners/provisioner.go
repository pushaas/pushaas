package provisioners

import (
	"github.com/pushaas/pushaas/pushaas/models"
)

type (
	PushServiceProvisionStatus   int
	PushServiceDeprovisionStatus int

	PushServiceProvisionResult struct {
		Instance *models.Instance
		EnvVars  map[string]string
		Status   PushServiceProvisionStatus
	}

	PushServiceDeprovisionResult struct {
		Instance *models.Instance
		Status   PushServiceDeprovisionStatus
	}

	PushServiceProvisioner interface {
		Provision(*models.Instance) *PushServiceProvisionResult
		Deprovision(*models.Instance) *PushServiceDeprovisionResult
	}
)

const (
	PushServiceProvisionStatusSuccess PushServiceProvisionStatus = iota
	PushServiceProvisionStatusFailure
)

const (
	PushServiceDeprovisionStatusSuccess PushServiceDeprovisionStatus = iota
	PushServiceDeprovisionStatusFailure
)

const EnvVarEndpoint = "PUSHAAS_ENDPOINT" // client apps use this var as the push-api endpoint
const EnvVarPassword = "PUSHAAS_PASSWORD" // client apps use this var as password to authenticate to push-api
const EnvVarUsername = "PUSHAAS_USERNAME" // client apps use this var as username to authenticate to push-api
