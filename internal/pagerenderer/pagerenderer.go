package pagerenderer

import (
	"board/internal/models"
	"board/internal/pagerenderer/upperapi"
	"board/internal/repo"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
)

const (
	BOARD_TMPL = "html/templates_for_dirty_front/board.html"
	PARTS_TMPL = "html/templates_for_dirty_front/parts.html"
	POSTS_TMPL = "html/templates_for_dirty_front/post.html"
	MAIN_TMPl  = "html/templates_for_dirty_front/main.html"
)

const (
	API  = "http://10.147.17.74:8080"
	SITE = "http://10.147.17.74:8081"

	CREATE = "/create"
)

type PageRenderer struct {
	Router *chi.Mux
}

func NewPageRenderer(pr *repo.Repo) *PageRenderer {
	r := chi.NewRouter()

	a := &PageRenderer{Router: r}

	r.Use(ValidateUser)

	r.Get("/{board}-{offset}-{n}", a.BoardPage)
	r.Get("/post/{id}-{offset}-{n}", a.PostPage)
	r.Get("/", a.MainPage)

	r.Post("/login", a.Login)
	r.Post("/quit", a.Quit)

	r.Route("/api", upperapi.UpperApi)

	FileServer(r, "/images", http.Dir("./images"))

	return a
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
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

	req, _ := http.NewRequest("GET", API+fmt.Sprintf("%v/get-%v-%v", board, offset, n), nil)
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

	var posts []models.Post = make([]models.Post, 0)
	data, err := io.ReadAll(rt.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = json.Unmarshal(data, &posts)
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
		User         string
	}

	var t = T{Board: board, Posts: posts, CreateAction: API + CREATE, Site: SITE,
		Offset: offset, User: r.Context().Value("Username").(string)}

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

	req, _ := http.NewRequest("GET", API+fmt.Sprintf("/get-%v", id), nil)
	req.Header.Add("JWT", r.Context().Value("JWT").(string))
	rt, err := http.DefaultClient.Do(req)
	//rt, err := http.Get(API + fmt.Sprintf("/get-%v", id))
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

	var op models.Post
	data, err := io.ReadAll(rt.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = json.Unmarshal(data, &op)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	req, _ = http.NewRequest("GET", API+fmt.Sprintf("/get-responses-%v-%v-%v", id, offset, n), nil)
	req.Header.Add("JWT", r.Context().Value("JWT").(string))
	rt, err = http.DefaultClient.Do(req)
	//rt, err = http.Get(API + fmt.Sprintf("/get-responses-%v-%v-%v", id, offset, n))
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

	var posts []models.Post = make([]models.Post, 0)
	data, err = io.ReadAll(rt.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = json.Unmarshal(data, &posts)
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
		User         string
	}

	var t = T{Id: id, Posts: posts, CreateAction: API + CREATE, Site: SITE,
		Offset: offset, Op: op, User: r.Context().Value("Username").(string)}

	err = ts.Execute(w, t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

func (pr PageRenderer) MainPage(w http.ResponseWriter, r *http.Request) {
	rt, err := http.Get(API + "/get-boards")
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

	var boards []models.Board
	data, err := io.ReadAll(rt.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = json.Unmarshal(data, &boards)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	type T struct {
		Boards []models.Board
		User   string
		API    string
	}

	var t = T{Boards: boards, User: r.Context().Value("Username").(string), API: API}

	ts, err := template.ParseFiles(MAIN_TMPl, PARTS_TMPL)
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

func (pr PageRenderer) Login(w http.ResponseWriter, r *http.Request) {
	user := models.User{}
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	err = schema.NewDecoder().Decode(&user, r.PostForm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	dataUrl := url.Values{}
	dataUrl.Add("Name", user.Name)
	dataUrl.Add("Pass", user.Pass)
	rt, err := http.Post(API+"/login", "application/x-www-form-urlencoded", strings.NewReader(dataUrl.Encode()))
	if rt.StatusCode != http.StatusOK || err != nil {
		rt, err := http.Post(API+"/create-user", "application/x-www-form-urlencoded", strings.NewReader(dataUrl.Encode()))
		if rt.StatusCode != http.StatusOK || err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error() + "\n" + rt.Status))
			return
		}
		rt, err = http.Post(API+"/login", "application/x-www-form-urlencoded", strings.NewReader(dataUrl.Encode()))
		if rt.StatusCode != http.StatusOK || err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error() + "\n" + rt.Status))
			return
		}
	}
	var mp map[string]string
	data, err := io.ReadAll(rt.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = json.Unmarshal(data, &mp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	tokenStr, ok := mp["JWT"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("go fuck yourself"))
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "JWT", Value: tokenStr})
	http.SetCookie(w, &http.Cookie{Name: "Username", Value: user.Name})
	http.Redirect(w, r, SITE, http.StatusFound)
}

func (pr PageRenderer) Quit(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "JWT", Value: "", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "Username", Value: "", MaxAge: -1})
	http.Redirect(w, r, SITE, http.StatusFound)
}

func ValidateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if username, err := r.Cookie("Username"); err == nil {
			ctx = context.WithValue(r.Context(), "Username", username.Value)
			tokenStr, _ := r.Cookie("JWT")
			ctx = context.WithValue(ctx, "JWT", tokenStr.Value)
		} else {
			ctx = context.WithValue(r.Context(), "Username", "")
			ctx = context.WithValue(ctx, "JWT", "")
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
