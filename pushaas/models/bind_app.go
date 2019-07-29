package models

type (
	BindApp struct {
		AppName    string   `json:"appName"`
		AppHost    string   `json:"appHost"`
	}
)

func BindAppFromForm(bindAppForm *BindAppForm) *BindApp {
	return &BindApp{
		AppName: bindAppForm.AppName,
		AppHost: bindAppForm.AppHost,
	}
}
