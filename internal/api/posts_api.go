package api

import (
	"board/internal/models"
	"board/internal/repo"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
)

const (
	MAX_MULTIPART_SIZE = 10485760
)

type Api struct {
	Router *chi.Mux
	Repo   *repo.Repo
}

func NewApi(pr *repo.Repo) *Api {
	r := chi.NewRouter()

	a := &Api{Router: r, Repo: pr}

	r.Use(a.ValidateUser)

	r.Post("/create", a.CreatePost)
	r.Get("/{board}/get-{offset}-{n}", a.GetNPosts)
	r.Get("/get-{id}", a.GetPost)
	r.Get("/get-responses-{id}-{offset}-{n}", a.GetResponses)

	r.Post("/create-user", a.CreateUser)
	r.Post("/login", a.Login)

	return a
}

var ErrEmptyText = fmt.Errorf("нельзя пустой текст")
var ErrWrongFile = fmt.Errorf("файл должен быть в Data")
var ErrBadMan = fmt.Errorf("ВЫ ПОСТИТЕ ПОД ЧУЖИМ ИМЕНЕМ")

func (a *Api) CreatePost(w http.ResponseWriter, r *http.Request) {
	var post models.Post
	err := r.ParseMultipartForm(MAX_MULTIPART_SIZE)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	err = schema.NewDecoder().Decode(&post, r.MultipartForm.Value)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	post.Timestamp = time.Now()
	if post.Text == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(ErrEmptyText.Error()))
		return
	}
	if post.Author == "" {
		post.Author = "Аноним"
	}
	if len(r.MultipartForm.File) != 0 {
		if _, ok := r.MultipartForm.File["Data"]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(ErrWrongFile.Error()))
			return
		}
		fr, err := r.MultipartForm.File["Data"][0].Open()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		data, err := io.ReadAll(fr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		post.Data = base64.StdEncoding.EncodeToString(data)
	}
	if post.ParentId != 0 {
		op, err := a.Repo.GetPost(post.ParentId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		post.Board = op.Board
		w.Write([]byte("Установлена доска " + op.Board + "\n"))
	}

	if name, ok := r.Context().Value("name").(string); ok {
		post.Author = name
	} else {
		_, err = a.Repo.GetUser(post.Author)
		if err == nil { //кто-то пытается анонимно постить под пользователя
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(ErrBadMan.Error()))
			return
		}
	}

	err = a.Repo.CreatePost(post)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("Ok"))
}

func (a *Api) GetNPosts(w http.ResponseWriter, r *http.Request) {
	offset, err := strconv.Atoi(chi.URLParam(r, "offset"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	n, err := strconv.Atoi(chi.URLParam(r, "n"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	board := "/" + chi.URLParam(r, "board")
	posts, err := a.Repo.GetNPostsFromBoard(n, offset, board, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	resultJson, err := json.Marshal(posts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(resultJson)
}

func (a *Api) GetPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	post, err := a.Repo.GetPost(int64(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	resultJson, err := json.Marshal(post)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(resultJson)
}

func (a *Api) GetResponses(w http.ResponseWriter, r *http.Request) {
	offset, err := strconv.Atoi(chi.URLParam(r, "offset"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	n, err := strconv.Atoi(chi.URLParam(r, "n"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	posts, err := a.Repo.GetResponsesForPost(id, offset, n, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	resultJson, err := json.Marshal(posts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(resultJson)
}
