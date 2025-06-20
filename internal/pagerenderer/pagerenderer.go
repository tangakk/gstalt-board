package pagerenderer

import (
	"board/internal/models"
	"board/internal/repo"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const (
	BOARD_TMPL = "html/templates_for_dirty_front/board.html"
	PARTS_TMPL = "html/templates_for_dirty_front/parts.html"
	POSTS_TMPL = "html/templates_for_dirty_front/post.html"
)

const (
	API  = "http://localhost:8080"
	SITE = "http://localhost:8081"

	CREATE = "/create"
)

type PageRenderer struct {
	PostsRepo *repo.Repo
	Router    *chi.Mux
}

func NewPageRenderer(pr *repo.Repo) *PageRenderer {
	r := chi.NewRouter()

	a := &PageRenderer{Router: r, PostsRepo: pr}

	r.Get("/{board}-{offset}-{n}", a.BoardPage)
	r.Get("/post/{id}-{offset}-{n}", a.PostPage)

	return a
}

func (pr PageRenderer) BoardPage(w http.ResponseWriter, r *http.Request) {
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

	posts, err := pr.PostsRepo.GetNPostsFromBoard(n, offset, board, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ts, err := template.ParseFiles(BOARD_TMPL, PARTS_TMPL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	type T struct {
		Board        string
		Posts        []models.Post
		CreateAction string
		Site         string
		Offset       int
	}

	var t = T{Board: board, Posts: posts, CreateAction: API + CREATE, Site: SITE,
		Offset: offset}

	err = ts.Execute(w, t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

func (pr PageRenderer) PostPage(w http.ResponseWriter, r *http.Request) {
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

	op, err := pr.PostsRepo.GetPost(int64(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	posts, err := pr.PostsRepo.GetResponsesForPost(id, offset, n, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ts, err := template.ParseFiles(POSTS_TMPL, PARTS_TMPL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	type T struct {
		Id           int
		Posts        []models.Post
		CreateAction string
		Site         string
		Offset       int
		Board        string
		Op           models.Post
	}

	var t = T{Id: id, Posts: posts, CreateAction: API + CREATE, Site: SITE,
		Offset: offset, Op: op}

	err = ts.Execute(w, t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}
