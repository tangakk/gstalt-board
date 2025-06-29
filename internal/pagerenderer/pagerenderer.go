package pagerenderer

import (
	"board/internal/models"
	"board/internal/pagerenderer/upperapi"
	"board/internal/repo"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	//API  = "http://10.147.17.74:8080"
	SITE = "/old"

	CREATE = "/old/create"
)

const (
	BASE_N = 20
)

type PagerendererConfig struct {
	API          string `env:"PAGERENDERER_API" env-default:"http://localhost:8080"`
	BASE_N       int    `env:"PAGERENDERER_POSTS_PER_PAGE" env-default:"20"`
	PORT         string `env:"PAGERENDERER_PORT" env-default:":8081"`
	DOMAIN       string `env:"DOMAIN" env-default:"gstalch.ru"`
	EXTRA_DOMAIN string `env:"EXTRA_DOMAIN" env-default:"gstalch.ru"`
	HTTPS        bool   `env:"HTTPS" env-default:"false"`
}

type PageRenderer struct {
	Router *chi.Mux
	PagerendererConfig
}

func NewPageRenderer(pr *repo.Repo, cfg PagerendererConfig) *PageRenderer {
	r := chi.NewRouter()

	a := &PageRenderer{Router: r, PagerendererConfig: cfg}

	r.Use(ValidateUser)

	r.Get("/", renderFile("html/templates/index.html"))
	r.Get("/{board}", renderFile("html/templates/board.html"))
	r.Get("/{board}/{post}", renderFile("html/templates/post.html"))

	r.Route("/old", func(r chi.Router) {
		//r.Use(ValidateUser)
		//r.Get("/{board}-{offset}-{n}", a.BoardPage)
		r.Get("/{board}", a.BoardPage)
		r.Get("/{board}/p{page}", a.BoardPage)
		//r.Get("/post/{id}-{offset}-{n}", a.PostPage)
		r.Get("/{board}/{id}", a.PostPage)
		r.Get("/", a.MainPage)
		r.Post("/create-board", a.CreateBoard)
		r.Post("/post", a.CreatePost)

		r.Post("/login", a.Login)
		r.Post("/quit", a.Quit)
	})
	ua := upperapi.UA{UpperApiConfig: upperapi.UpperApiConfig{NORMAL_API: a.API}}

	r.Route("/api", ua.UpperApi)

	FileServer(r, "/images", http.Dir("./images"))
	FileServer(r, "/internal", http.Dir("./html/templates"))

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
	/*offset, err := strconv.Atoi(chi.URLParam(r, "offset"))
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
	}*/
	pageParam := chi.URLParam(r, "page")
	page, err := strconv.Atoi(pageParam)
	if pageParam != "" {
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
	}
	board := "/" + chi.URLParam(r, "board")

	req, _ := http.NewRequest("GET", pr.API+fmt.Sprintf("%v/get-recent-%v-%v", board, page*BASE_N, BASE_N), nil)
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
	ts, err := template.New("board.html").Funcs(template.FuncMap{"StringTime": unixToString, "IsMP4": isMP4}).
		ParseFiles(BOARD_TMPL, PARTS_TMPL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	rt, err = http.Get(pr.API + "/get-boards")
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
	data, err = io.ReadAll(rt.Body)
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

	type P struct {
		models.Post
		ResponsesPosts []models.Post
	}

	new_posts := make([]P, len(posts))
	for i, p := range posts {
		req, _ = http.NewRequest("GET", pr.API+fmt.Sprintf("%v/get-responses-%v-%v-%v/r", board, p.Id, 0, 3), nil)
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

		var resps []models.Post
		data, err = io.ReadAll(rt.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		err = json.Unmarshal(data, &resps)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		for i, v := range resps {
			resps[i].Text = template.HTML(makeResponsesLinks(v, r, pr.API))
		}
		//slices.Reverse(resps)
		p.Text = template.HTML(makeResponsesLinks(p, r, pr.API))
		new_posts[i] = P{Post: p, ResponsesPosts: resps}
	}

	type T struct {
		Board        string
		Posts        []P
		CreateAction string
		Site         string
		Page         int
		User         string
		Boards       []models.Board
		Id           int
	}

	var t = T{Board: board, Posts: new_posts, CreateAction: CREATE, Site: "boardPage",
		Page: page, User: r.Context().Value("Username").(string), Boards: boards, Id: 0}

	err = ts.Execute(w, t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

func (pr PageRenderer) PostPage(w http.ResponseWriter, r *http.Request) {
	/*offset, err := strconv.Atoi(chi.URLParam(r, "offset"))

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
	}*/
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	board := chi.URLParam(r, "board")

	req, _ := http.NewRequest("GET", pr.API+fmt.Sprintf("/%v/get-one-%v", board, id), nil)
	//fmt.Println(err)
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

	op.Text = template.HTML(makeResponsesLinks(op, r, pr.API))

	req, _ = http.NewRequest("GET", pr.API+fmt.Sprintf("/%v/get-responses-%v-%v-%v", board, id, 0, op.Responses), nil)
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

	for i, p := range posts {
		posts[i].Text = template.HTML(makeResponsesLinks(p, r, pr.API))
	}

	ts, err := template.New("post.html").Funcs(template.FuncMap{"StringTime": unixToString, "IsMP4": isMP4}).
		ParseFiles(POSTS_TMPL, PARTS_TMPL)
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
		Answer       bool
	}

	var t = T{Id: id, Posts: posts, CreateAction: CREATE, Site: SITE,
		Offset: 0, Op: op, User: r.Context().Value("Username").(string), Board: "/" + board}

	err = ts.Execute(w, t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

func (pr PageRenderer) MainPage(w http.ResponseWriter, r *http.Request) {
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

	var t = T{Boards: boards, User: r.Context().Value("Username").(string), API: pr.API}

	ts, err := template.New("main.html").Funcs(template.FuncMap{"StringTime": unixToString, "IsMP4": isMP4}).
		ParseFiles(MAIN_TMPl, PARTS_TMPL)
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
	rt, err := http.Post(pr.API+"/login", "application/x-www-form-urlencoded", strings.NewReader(dataUrl.Encode()))
	if rt.StatusCode != http.StatusOK || err != nil {
		rt, err = http.Post(pr.API+"/create-user", "application/x-www-form-urlencoded", strings.NewReader(dataUrl.Encode()))
		if rt.StatusCode != http.StatusOK || err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error() + "\n" + rt.Status))
			return
		}
		rt, err = http.Post(pr.API+"/login", "application/x-www-form-urlencoded", strings.NewReader(dataUrl.Encode()))
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
		w.Write([]byte(err.Error() + ":" + string(data)))
		return
	}

	tokenStr, ok := mp["JWT"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("go fuck yourself"))
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "JWT", Value: tokenStr, Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: "Username", Value: base64.StdEncoding.EncodeToString([]byte(user.Name)), Path: "/"})
	http.Redirect(w, r, SITE, http.StatusFound)
}

func (pr PageRenderer) CreateBoard(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequest("POST", pr.API+"/create-board", bytes.NewReader(body))
	req.Header = r.Header
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		http.Error(w, string(data), resp.StatusCode)
		return
	}
	//var boards []models.Board
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(data)
}

func (pr PageRenderer) CreatePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequest("POST", pr.API+"/create", bytes.NewReader(body))
	req.Header = r.Header
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		http.Error(w, string(data), resp.StatusCode)
		return
	}
	//var boards []models.Board
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(data)
}

func (pr PageRenderer) Quit(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "JWT", Value: "", MaxAge: -1, Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: "Username", Value: "", MaxAge: -1, Path: "/"})
	http.Redirect(w, r, SITE, http.StatusFound)
}

func ValidateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if username, err := r.Cookie("Username"); err == nil && username.Path == "" {
			v, _ := base64.StdEncoding.DecodeString(username.Value)
			ctx = context.WithValue(r.Context(), "Username", string(v))
			tokenStr, _ := r.Cookie("JWT")
			ctx = context.WithValue(ctx, "JWT", tokenStr.Value)
		} else {
			ctx = context.WithValue(r.Context(), "Username", "")
			ctx = context.WithValue(ctx, "JWT", "")
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func renderFile(file string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, _ := os.Open(file)
		data, _ := io.ReadAll(file)
		w.Write(data)
	}
}

func unixToString(u int64) string {
	t := time.Unix(u, 0)
	return t.Format("2006/01/02    15:04")
}

func isMP4(f string) bool {
	if f[len(f)-4:] == ".mp4" {
		return true
	} else {
		return false
	}
}

var reg = regexp.MustCompile(">>[0-9]+")

func makeResponsesLinks(post models.Post, r *http.Request, API string) string {
	s := string(post.Text)
	var res string
	matches := reg.FindAllIndex([]byte(s), -1)
	//addr := r.URL.String()
	if len(matches) == 0 {
		return s
	}
	var bs = []byte(s)
	for i, v := range matches {
		ms := string(bs[v[0]:v[1]])
		link := ms
		if id1, _ := strconv.Atoi(ms); id1 > int(post.Id) {
			continue
		}

		req, _ := http.NewRequest("GET", API+fmt.Sprintf("%v/get-one-%v", post.Board, ms[2:]), nil)
		//fmt.Println(err)
		req.Header.Add("JWT", r.Context().Value("JWT").(string))
		rt, err := http.DefaultClient.Do(req)
		//rt, err := http.Get(API + fmt.Sprintf("/get-%v", id))
		if err == nil {
			if rt.StatusCode == http.StatusOK {
				var op models.Post
				data, err := io.ReadAll(rt.Body)
				if err == nil {
					err = json.Unmarshal(data, &op)
					if err == nil {
						switch {
						case op.ParentId == 0:
							link = fmt.Sprintf("<a href=\"/old%v/%v\">%v</a>", op.Board, ms[2:], ms)
						default:
							link = fmt.Sprintf("<a href=\"/old%v/%v#%v\">%v</a>", op.Board, op.ParentId, ms[2:], ms)
						}
					}
				}
			}
		}
		//fmt.Println(req.URL)

		var add string
		if i == 0 {
			add = string(bs[:v[0]])
		} else {
			add = string(bs[matches[i-1][1]:v[0]])
		}
		res += html.EscapeString(add) + link
	}
	res += html.EscapeString(string(bs[matches[len(matches)-1][1]:]))
	return res
}
