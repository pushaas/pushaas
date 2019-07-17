package models

type (
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Fields  string `json:"fields"`
	}
)

type ErrorCode int

const (
	/*
		instance
	*/
	ErrorInstanceCreateFailed        = 10
	ErrorInstanceCreateAlreadyExists = 11
	ErrorInstanceCreateInvalidPlan   = 12

	ErrorInstanceDeleteFailed = 20

	/*
		bind
	*/
)
