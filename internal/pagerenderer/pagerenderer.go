package pagerenderer

import (
	"board/internal/models"
	"board/internal/postsrepo"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const (
	BOARD_TEMPLATE = "html/templates/board.html"
)

const (
	API  = "http://localhost:8080"
	SITE = "http://localhost:8081"

	CREATE = "/create"
)

type PageRenderer struct {
	PostsRepo *postsrepo.PostsRepo
	Router    *chi.Mux
}

func NewPageRenderer(pr *postsrepo.PostsRepo) *PageRenderer {
	r := chi.NewRouter()

	a := &PageRenderer{Router: r, PostsRepo: pr}

	r.Get("/{board}-{offset}-{n}", a.BoardPage)

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

	posts, err := pr.PostsRepo.GetNPostsFromBoard(n, offset, board)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ts, err := template.ParseFiles(BOARD_TEMPLATE)
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
