package models

import (
	"github.com/go-bongo/bongo"
)

type Instance struct {
	bongo.DocumentBase `bson:",inline"`
	Description        string `json:"description"`
}
