package upperapi

import (
	"board/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UpperApiConfig struct {
	NORMAL_API string `env:"UPPERAPI_NORMAL_API_ADDR" env-default:"http://localhost:8080"`
}

type UA struct {
	UpperApiConfig
}

func (ua *UA) UpperApi(r chi.Router) {
	r.Get("/get/boards", ua.getBoards)
	r.Get("/get/posts", ua.getPosts)
	r.Get("/get/comments", ua.getComments)
	r.Get("/whoami", ua.getMe)
	r.Post("/post/post", ua.post)
	r.Post("/post/board", ua.postBoard)
}

var ErrBadRequest = fmt.Errorf("херовый реквест")
var ErrNoResponseFromNormalApi = fmt.Errorf("нормальное api не отвечает")

func (ua *UA) getBoards(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("GET", ua.NORMAL_API+"/get-boards", nil)
	if err != nil {
		http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, ErrNoResponseFromNormalApi.Error(), http.StatusInternalServerError)
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
	/*
		err = json.Unmarshal(data, &boards)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		var result = []any{}
		for _, board := range boards {
			result = append(result, map[string]string{
				"Name":        board.Name,
				"Description": board.Description,
				"url":         r.Host + board.Name,
			})
		}
		resultJson, err := json.Marshal(boards)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}*/
	w.Write(data)
}

func (ua *UA) getMe(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("GET", ua.NORMAL_API+"/whoami", nil)
	if err != nil {
		http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, ErrNoResponseFromNormalApi.Error(), http.StatusInternalServerError)
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

func (ua *UA) getPosts(w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
		return
	}
	board := params.Get("board")
	var req *http.Request
	if params.Get("from") == "" {
		req, err = http.NewRequest("GET", ua.NORMAL_API+"/"+board+"/get-all", nil)
	} else {
		offset, err := strconv.Atoi(params.Get("from"))
		if err != nil {
			http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
			return
		}
		if offset < 0 {
			offset = 0
		}
		to, err := strconv.Atoi(params.Get("to"))
		if err != nil {
			http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
			return
		}
		n := to - offset + 1
		if params.Get("last") == "1" {
			n = -n
		}
		req, err = http.NewRequest("GET", fmt.Sprintf("%v/%v/get-%v-%v", ua.NORMAL_API, board, offset, -n), nil)
	}
	if err != nil {
		http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, ErrNoResponseFromNormalApi.Error(), http.StatusInternalServerError)
		return
	}
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		http.Error(w, string(data), resp.StatusCode)
		return
	}
	if params.Get("last") == "1" {
		var res []models.Post
		json.Unmarshal(data, &res)
		slices.Reverse(res)
		data, err = json.Marshal(res)
		if err != nil {
			http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Write(data)
}

func (ua *UA) getComments(w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
		return
	}
	var req *http.Request
	offset, err := strconv.Atoi(params.Get("from"))
	if err != nil {
		http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
		return
	}
	if offset < 0 {
		offset = 0
	}
	to, err := strconv.Atoi(params.Get("to"))
	if err != nil {
		http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
		return
	}
	id, err := strconv.Atoi(params.Get("post"))
	if err != nil {
		http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
		return
	}
	board := params.Get("board")
	n := to - offset
	if params.Get("last") != "1" {
		req, err = http.NewRequest("GET", fmt.Sprintf("%v/%v/get-responses-%v-%v-%v", ua.NORMAL_API, board, id, offset, n), nil)
	} else {
		req, err = http.NewRequest("GET", fmt.Sprintf("%v/%v/get-responses-%v-%v-%v/r", ua.NORMAL_API, board, id, offset, n), nil)
	}
	if err != nil {
		http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, ErrNoResponseFromNormalApi.Error(), http.StatusInternalServerError)
		return
	}
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		http.Error(w, string(data), resp.StatusCode)
		return
	}
	w.Write(data)
}

func (ua *UA) post(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequest("POST", ua.NORMAL_API+"/create", bytes.NewReader(body))
	req.Header = r.Header
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, ErrNoResponseFromNormalApi.Error(), http.StatusInternalServerError)
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

func (ua *UA) postBoard(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequest("POST", ua.NORMAL_API+"/create-board", bytes.NewReader(body))
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
