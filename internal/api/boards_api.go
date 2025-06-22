package api

import (
	"board/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
)

var ErrNotAdmin = fmt.Errorf("только для админов")
var ErrInvalidBoardName = fmt.Errorf("/onlylatinand0123456789inboardname")

func (a *Api) CreateBoard(w http.ResponseWriter, r *http.Request) {
	var board models.Board
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	err = schema.NewDecoder().Decode(&board, r.PostForm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	admins := strings.Split(r.PostForm.Get("Admins"), ",")
	board.Admins = admins
	users := strings.Split(r.PostForm.Get("UsersList"), ",")
	board.UsersList = users
	user, _ := r.Context().Value("user").(models.User)
	if !user.Admin {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(ErrNotAdmin.Error()))
		return
	}
	board.Owner = user.Name
	if !validBoardName(board.Name) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(ErrInvalidBoardName.Error()))
		return
	}
	err = a.Repo.CreateBoard(board)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("Ok"))
}

func (a *Api) GetAllBoards(w http.ResponseWriter, r *http.Request) {
	boards, err := a.Repo.GetAllBoards()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	user, _ := r.Context().Value("user").(models.User)
	if !user.Admin {
		for i := 0; i < len(boards); i++ {
			boards[i].Admins = []string{}
			boards[i].UsersList = []string{}
		}
	}
	resultJson, err := json.Marshal(boards)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(resultJson)
}

func (a *Api) GetBoard(w http.ResponseWriter, r *http.Request) {
	boardStr := "/" + chi.URLParam(r, "board")
	board, err := a.Repo.GetBoard(boardStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	user, _ := r.Context().Value("user").(models.User)
	if !user.Admin && !slices.Contains(board.Admins, user.Name) {
		board.Admins = []string{}
		board.UsersList = []string{}
	}
	resultJson, err := json.Marshal(board)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(resultJson)
}

func (a *Api) UpdateBoard(w http.ResponseWriter, r *http.Request) {
	var board models.Board
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	err = schema.NewDecoder().Decode(&board, r.PostForm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	admins := strings.Split(r.PostForm.Get("Admins"), ",")
	board.Admins = admins
	users := strings.Split(r.PostForm.Get("UsersList"), ",")
	board.UsersList = users

	boardInDb, err := a.Repo.GetBoard(board.Name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	user, _ := r.Context().Value("user").(models.User)
	if !user.Admin && user.Name != boardInDb.Owner && !slices.Contains(boardInDb.Admins, user.Name) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(ErrNotAdmin.Error()))
		return
	}
	if !user.Admin && user.Name != boardInDb.Owner { //админ борды, но не владелец/суперадмин
		board.Admins = boardInDb.Admins //не даём менять админов
		board.Mode = boardInDb.Mode     //и мод
	}
	err = a.Repo.UpdateBoard(board)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("Ok"))
}

func validBoardName(str string) bool {
	return regexp.MustCompile(`^\/[a-z0-9]+$`).MatchString(str)
}
