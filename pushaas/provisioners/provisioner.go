package provisioners

type (
	Provisioner interface {
		Provision(string) error
		Deprovision(string) error
	}
)
