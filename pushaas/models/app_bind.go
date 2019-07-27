package models

type (
	AppBind struct {
		AppName    string   `json:"appName"`
		AppHost    string   `json:"appHost"`
	}
)

func AppBindFromForm(bindAppForm *BindAppForm) *AppBind {
	return &AppBind{
		AppName: bindAppForm.AppName,
		AppHost: bindAppForm.AppHost,
	}
}
