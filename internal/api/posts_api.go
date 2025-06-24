package api

import (
	"board/internal/models"
	"board/internal/models/boardmodes"
	"board/internal/repo"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
)

const (
	MAX_MULTIPART_SIZE = 10485760
	ANON               = "Аноним"
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
	r.Get("/{board}/get-{offset}-{n}/{all}", a.GetNPosts)
	r.Get("/{board}/get-all", a.GetAllPosts)
	r.Get("/{board}/get-all/{all}", a.GetAllPosts)
	r.Get("/get-{id}", a.GetPost)
	r.Get("/get-responses-{id}-{offset}-{n}", a.GetResponses)
	r.Get("/get-all-responses-{id}", a.GetAllResponses)

	r.Post("/create-user", a.CreateUser)
	r.Post("/login", a.Login)
	r.Post("/op", a.Op)

	r.Post("/create-board", a.CreateBoard)
	r.Get("/get-boards", a.GetAllBoards)
	r.Get("/{board}/get-board", a.GetBoard)
	r.Post("/update-board", a.UpdateBoard)

	return a
}

var ErrEmptyText = fmt.Errorf("нельзя пустой текст")
var ErrWrongFile = fmt.Errorf("файл должен быть в Data")
var ErrBadMan = fmt.Errorf("ВЫ ПОСТИТЕ ПОД ЧУЖИМ ИМЕНЕМ")
var ErrNoBoard = fmt.Errorf("такой доски не существует")
var ErrCantPost = fmt.Errorf("вы не можете постить на этой доске")

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
	post.Timestamp = time.Now().Unix()
	if post.Text == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(ErrEmptyText.Error()))
		return
	}
	if post.Author == "" {
		post.Author = ANON
	}
	post.Author = post.Author[:min(MAX_NAME_LEN, len(post.Author))]
	if strings.Contains(post.Author, ",") {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(ErrForbiddenChars.Error()))
		return
	}
	if len(r.MultipartForm.File) != 0 {
		if _, ok := r.MultipartForm.File["Data"]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(ErrWrongFile.Error()))
			return
		}
		file, handler, err := r.FormFile("Data")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()
		dst, err := os.Create("images/" + strconv.FormatInt(time.Now().Unix(), 10) + "." + handler.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		post.Data = dst.Name()
	}
	if post.ParentId != 0 {
		op, err := a.Repo.GetPost(post.ParentId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		post.Board = op.Board
		//w.Write([]byte("Установлена доска " + op.Board + "\n"))
	}
	user, _ := r.Context().Value("user").(models.User)

	if user.Name != "" {
		post.Author = user.Name
	} else {
		_, err = a.Repo.GetUser(post.Author)
		if err == nil {
			//кто-то пытается анонимно постить под пользователя
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(ErrBadMan.Error()))
			return
		} else {
			user.Name = ANON
		}
	}

	board, err := a.Repo.GetBoard(post.Board)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(ErrNoBoard.Error()))
		return
	}

	if !userCanPostOnBoard(user, board) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(ErrCantPost.Error()))
		return
	}

	err = a.Repo.CreatePost(post)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("Ok"))
}

var ErrCantRead = fmt.Errorf("вы не можете читать эту доску")

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
	boardStr := "/" + chi.URLParam(r, "board")
	board, err := a.Repo.GetBoard(boardStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(ErrNoBoard.Error()))
		return
	}

	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		user.Name = ANON
	}
	if !userCanReadBoard(user, board) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(ErrCantRead.Error()))
		return
	}

	all := chi.URLParam(r, "all") == "all"

	reverse := true
	if n < 0 {
		n = -n
		reverse = false
	}
	posts, err := a.Repo.GetNPostsFromBoard(n, offset, boardStr, reverse, all)
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

func (a *Api) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	boardStr := "/" + chi.URLParam(r, "board")
	board, err := a.Repo.GetBoard(boardStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(ErrNoBoard.Error()))
		return
	}

	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		user.Name = ANON
	}
	if !userCanReadBoard(user, board) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(ErrCantRead.Error()))
		return
	}

	all := chi.URLParam(r, "all") == "all"

	posts, err := a.Repo.GetAllPostsIdFromBoard(boardStr, true, all)
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

	board, err := a.Repo.GetBoard(post.Board)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(ErrNoBoard.Error()))
		return
	}

	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		user.Name = ANON
	}
	if !userCanReadBoard(user, board) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(ErrCantRead.Error()))
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

	op, _ := a.Repo.GetPost(int64(id))

	board, err := a.Repo.GetBoard(op.Board)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(ErrNoBoard.Error()))
		return
	}

	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		user.Name = ANON
	}
	if !userCanReadBoard(user, board) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(ErrCantRead.Error()))
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

func (a *Api) GetAllResponses(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	op, _ := a.Repo.GetPost(int64(id))

	board, err := a.Repo.GetBoard(op.Board)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(ErrNoBoard.Error()))
		return
	}

	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		user.Name = ANON
	}
	if !userCanReadBoard(user, board) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(ErrCantRead.Error()))
		return
	}

	posts, err := a.Repo.GetAllResponsesForPost(id, false)
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

func userCanPostOnBoard(user models.User, board models.Board) bool {
	if user.Admin {
		return true
	}
	switch board.Mode {
	case boardmodes.OPEN:
		return true
	case boardmodes.LIGHTBLACK:
		fallthrough
	case boardmodes.BLACK:
		if slices.Contains(board.UsersList, user.Name) {
			return false
		} else {
			return true
		}
	case boardmodes.LIGHTWHITE:
		fallthrough
	case boardmodes.WHITE:
		if slices.Contains(board.UsersList, user.Name) {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}

func userCanReadBoard(user models.User, board models.Board) bool {
	if user.Admin {
		return true
	}
	switch board.Mode {
	case boardmodes.OPEN:
		return true
	case boardmodes.LIGHTBLACK:
		return true
	case boardmodes.LIGHTWHITE:
		return true
	case boardmodes.BLACK:
		if slices.Contains(board.UsersList, user.Name) {
			return false
		} else {
			return true
		}
	case boardmodes.WHITE:
		if slices.Contains(board.UsersList, user.Name) {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}
