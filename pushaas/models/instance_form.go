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
	InstanceFormInvalid
)

func (i *InstanceForm) Validate() InstanceFormValidation {
	if i.Plan != PlanSmall {
		return InstanceFormInvalid
	}

	if i.Name == "" {
		return InstanceFormInvalid
	}

	return InstanceFormValid
}
