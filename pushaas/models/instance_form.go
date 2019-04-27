package models

type (
	InstanceFormValidation int

	InstanceForm struct {
		Name string
		Plan string
		Team string
		User string
	}
)

const (
	InstanceFormValid InstanceFormValidation = iota
	InstanceFormInvalidPlan
)

func (i *InstanceForm) Validate() InstanceFormValidation {
	if i.Plan != PlanSmall {
		return InstanceFormInvalidPlan
	}

	// TODO validate fields

	return InstanceFormValid
}
