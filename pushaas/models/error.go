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
	ErrorInstanceRetrievalFailed   = 10
	ErrorInstanceRetrievalNotFound = 11

	ErrorInstanceCreateFailed                  = 20
	ErrorInstanceCreateDispatchProvisionFailed = 21
	ErrorInstanceCreateAlreadyExists           = 22
	ErrorInstanceCreateInvalidData             = 23

	ErrorInstanceDeleteFailed                    = 30
	ErrorInstanceDeleteDispatchDeprovisionFailed = 31
	ErrorInstanceDeleteNotFound                  = 32

	ErrorInstanceStatusRetrievalFailed   = 40
	ErrorInstanceStatusRetrievalNotFound = 41

	/*
		bind
	*/
	ErrorBindAppNotFound        = 100
	ErrorBindAppAlreadyBound    = 101
	ErrorBindAppFailed          = 102
	ErrorBindAppInstancePending = 103
	ErrorBindAppInstanceFailed  = 104

	ErrorUnbindAppNotFound = 110
	ErrorUnbindAppNotBound = 111
	ErrorUnbindAppFailed   = 112
)
