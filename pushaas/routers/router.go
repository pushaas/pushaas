package routers

type (
	Router interface{}

	Response struct {
		Data  interface{} `json:"data"`
		Error string      `json:"error,omitempty"`
	}
)
