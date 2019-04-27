package models

import (
	"github.com/go-bongo/bongo"
)

type (
	InstanceStatus string

	InstanceBinding struct {
		AppName    string   `json:"appName"`
		AppHost    string   `json:"appHost"`
		UnitsHosts []string `json:"unitsHosts"`
	}

	Instance struct {
		bongo.DocumentBase `bson:",inline"`
		Name               string            `json:"name"`
		Plan               string            `json:"plan"`
		Team               string            `json:"team"`
		User               string            `json:"user"`
		Status             InstanceStatus    `json:"status"`
		Bindings           []InstanceBinding `json:"bindings"`
	}
)

const (
	InstanceStatusRunning = InstanceStatus("running")
	InstanceStatusPending = InstanceStatus("pending")
	InstanceStatusFailed  = InstanceStatus("failed")
)
