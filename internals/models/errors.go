package models

import "errors"

var ErrInvalidNumberArgument = errors.New("invalid number of arguments")
var ErrInvalidCode = errors.New("invalid code")
var ErrInvalidName = errors.New("invalid name")
var ErrInvalidMap = errors.New("invalid map")
var ErrInvalidMode = errors.New("invalid mode")
var ErrRoomAlreadyExist = errors.New("room already exist")
