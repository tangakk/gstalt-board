package models

import (
	"board/internal/models/boardmodes"
	"time"
)

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

type Board struct {
	Name        string
	Description string
	Admins      []string
	Mode        boardmodes.BoardMode
	UsersList   []string
	Owner       string
}
