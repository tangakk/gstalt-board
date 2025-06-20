package api

import (
	"board/internal/models"
	"board/internal/postsrepo"
	"encoding/base64"
	"encoding/json"
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
	Router    *chi.Mux
	PostsRepo *postsrepo.PostsRepo
}

func NewApi(pr *postsrepo.PostsRepo) *Api {
	r := chi.NewRouter()

	a := &Api{Router: r, PostsRepo: pr}

	r.Post("/create", a.CreatePost)
	r.Get("/{board}/get-{offset}-{n}", a.GetNPosts)

	return a
}

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
		w.Write([]byte("text can't be empty"))
		return
	}
	if post.Author == "" {
		post.Author = "Аноним"
	}
	if len(r.MultipartForm.File) != 0 {
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
	err = a.PostsRepo.CreatePost(post)
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
	posts, err := a.PostsRepo.GetNPostsFromBoard(n, offset, board)
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
