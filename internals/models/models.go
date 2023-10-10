package models

import (
	"errors"
	"time"
)

type Room struct {
	Code       string    `json:"code"`
	Mode       string    `json:"name"`
	Hoster     string    `json:"hoster"`
	Map        string    `json:"map"`
	Descrition string    `json:"description"`
	Time       time.Time `json:"time"`
}

var ErrInvalidNumberArgument = errors.New("invalid number of arguments")
var ErrInvalidCode = errors.New("invalid code")
var ErrInvalidName = errors.New("invalid name")
var ErrInvalidMap = errors.New("invalid map")
var ErrInvalidMode = errors.New("invalid mode")
