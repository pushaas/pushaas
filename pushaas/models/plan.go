package models

const (
	PlanSmall = "small"
)

type Plan struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
