package models

import (
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

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Hoster struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Followers []User    `json:"subs"`
	LastSend  time.Time `json:"last_send"`
}

type Follower struct {
	ID      int64  `json:"id"`
	Hosters []User `json:"hosters"`
}

type UserList struct {
	Users []User `json:"users"`
}
