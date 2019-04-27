package models

type (
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Fields  string `json:"fields"`
	}
)
