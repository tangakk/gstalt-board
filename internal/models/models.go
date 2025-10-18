package models

import (
	"board/internal/models/boardmodes"
	"html/template"
)

type Post struct {
	Id           int64
	Author       string
	Text         template.HTML
	Timestamp    int64
	Data         string
	ParentId     int64
	Board        string
	Responses    int64
	LastAnswered int64
}

type User struct {
	Name  string
	Pass  string
	Admin bool
}

type Board struct {
	Name                    string
	Description             string
	Admins                  []string
	Mode                    boardmodes.BoardMode
	UsersList               []string
	Owner                   string
	PostsCount              int64
	PostsCountWithResponses int64
	Image                   string
}
