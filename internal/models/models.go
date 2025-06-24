package models

import (
	"board/internal/models/boardmodes"
)

type Post struct {
	Id        int64
	Author    string
	Text      string
	Timestamp int64
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
