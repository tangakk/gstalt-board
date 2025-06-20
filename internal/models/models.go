package models

import "time"

type Post struct {
	Id        int64
	Author    string
	Text      string
	Timestamp time.Time
	Data      string
	ParentId  int64
	Board     string
}

type User struct {
	Name  string
	Pass  string
	Admin bool
}
