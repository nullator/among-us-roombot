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
	ID         int64     `json:"id"`
	Warning    bool      `json:"status"`
}

type RoomList []Room

func (r RoomList) Len() int {
	return len(r)
}

func (r RoomList) Less(i, j int) bool {
	return r[i].Time.Before(r[j].Time)
}

func (r RoomList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

var ErrInvalidNumberArgument = errors.New("invalid number of arguments")
var ErrInvalidCode = errors.New("invalid code")
var ErrInvalidName = errors.New("invalid name")
var ErrInvalidMap = errors.New("invalid map")
var ErrInvalidMode = errors.New("invalid mode")
var ErrRoomAlreadyExist = errors.New("room already exist")
