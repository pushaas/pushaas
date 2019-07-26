package models

import (
	"encoding/json"
)

const (
	InstanceStatusPending = InstanceStatus("pending")
	InstanceStatusRunning = InstanceStatus("running")
	InstanceStatusFailed  = InstanceStatus("failed")
)

type (
	InstanceStatus string

	InstanceBinding struct {
		AppName    string   `json:"appName"`
		AppHost    string   `json:"appHost"`
		// TODO
		//UnitsHosts []string `json:"unitsHosts"`
	}

	Instance struct {
		Name               string            `json:"name"`
		Plan               string            `json:"plan"`
		Team               string            `json:"team"`
		User               string            `json:"user"`
		Status             InstanceStatus    `json:"status"`
		// TODO
		//Bindings           []InstanceBinding `json:"bindings"`
	}
)

func (i InstanceStatus) MarshalBinary() ([]byte, error) {
	return []byte(i), nil
}

func (i InstanceStatus) UnmarshalBinary(data []byte) error{
	return json.Unmarshal(data, i)
}

func InstanceFromInstanceForm(instanceForm *InstanceForm) *Instance {
	return &Instance{
		Name: instanceForm.Name,
		Plan: instanceForm.Plan,
		Team: instanceForm.Team,
		User: instanceForm.User,
	}
}

