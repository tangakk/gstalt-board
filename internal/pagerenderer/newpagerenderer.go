package pagerenderer

import (
	"board/internal/models"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const (
	NEW_MAIN_TMPL  = "html/templates/index.html"
	NEW_BOARD_TMPL = "html/templates/board.html"
)

func (pr *PageRenderer) NewMainPage(w http.ResponseWriter, r *http.Request) {
	type T struct {
		Title        string
		IsRegistered bool
		Username     string
		Avatar       string
		Boards       []struct {
			models.Board
			IsSuperadmin bool
		}
	}
	var t = T{}
	t.Title = "ГЕШТАЛЬЧ"
	if r.Context().Value("Username").(string) != "" {
		t.IsRegistered = true
		t.Username = r.Context().Value("Username").(string)
	}
	rt, err := http.Get(pr.API + "/get-boards")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if rt.StatusCode != http.StatusOK {
		w.WriteHeader(rt.StatusCode)
		w.Write([]byte(rt.Status))
		return
	}
	data, err := io.ReadAll(rt.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = json.Unmarshal(data, &t.Boards)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	rt, err = http.Get(pr.API + "/whoami")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if rt.StatusCode != http.StatusOK {
		w.WriteHeader(rt.StatusCode)
		w.Write([]byte(rt.Status))
		return
	}
	data, err = io.ReadAll(rt.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	var user models.User
	err = json.Unmarshal(data, &user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	for i := range t.Boards {
		t.Boards[i].IsSuperadmin = user.Admin
	}
	ts, err := template.New("index.html").ParseFiles(NEW_MAIN_TMPL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = ts.Execute(w, t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

func (pr *PageRenderer) NewBoardPage(w http.ResponseWriter, r *http.Request) {
	type T struct {
		Title        string
		IsRegistered bool
		Username     string
		Avatar       string
		Posts        []struct {
			models.Post
			CanDelete bool
		}
	}
	var t = T{}
	if r.Context().Value("Username").(string) != "" {
		t.IsRegistered = true
		t.Username = r.Context().Value("Username").(string)
	}

	t.Title = "/" + chi.URLParam(r, "board")

	page := 0 //TODO: нормальные страницы

	req, _ := http.NewRequest("GET", pr.API+fmt.Sprintf("%v/get-recent-%v-%v", t.Title, page*BASE_N, BASE_N), nil)
	req.Header.Add("JWT", r.Context().Value("JWT").(string))
	rt, err := http.DefaultClient.Do(req)
	//rt, err := http.Get(API + fmt.Sprintf("%v/get-%v-%v", board, offset, n))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if rt.StatusCode != http.StatusOK {
		w.WriteHeader(rt.StatusCode)
		w.Write([]byte(rt.Status))
		return
	}

	data, err := io.ReadAll(rt.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = json.Unmarshal(data, &t.Posts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	rt, err = http.Get(pr.API + "/whoami")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if rt.StatusCode != http.StatusOK {
		w.WriteHeader(rt.StatusCode)
		w.Write([]byte(rt.Status))
		return
	}
	data, err = io.ReadAll(rt.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	var user models.User
	err = json.Unmarshal(data, &user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	for i := range t.Posts {
		t.Posts[i].CanDelete = user.Admin || (t.Posts[i].Author == t.Username)
	}
	ts, err := template.New("board.html").ParseFiles(NEW_BOARD_TMPL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = ts.Execute(w, t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}
