package models

type (
	InstanceForm struct {
		Name string
		Plan string
		Team string
		User string
	}
)

func (i *InstanceForm) Validate() error {
	// TODO implement
	return nil
}
