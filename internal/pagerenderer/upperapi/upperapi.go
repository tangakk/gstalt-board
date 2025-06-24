package upperapi

import (
	"board/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const NORMAL_API = "http://localhost:8080"

func UpperApi(r chi.Router) {
	r.Get("/get/boards", getBoards)
	r.Get("/get/posts", getPosts)
	r.Get("/get/comments", getComments)
}

var ErrBadRequest = fmt.Errorf("{\"Error_desc\":\"херовый реквест\"}")
var ErrNoResponseFromNormalApi = fmt.Errorf("{\"Error_desc\":\"нормальное api не отвечает\"}")

func getBoards(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("GET", NORMAL_API+"/get-boards", nil)
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
	var boards []models.Board
	data, err := io.ReadAll(resp.Body)
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
	var result = []any{}
	for _, board := range boards {
		result = append(result, map[string]string{
			"Name":        board.Name,
			"Description": board.Description,
			"url":         r.Host + board.Name,
		})
	}
	resultJson, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(resultJson)
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
		return
	}
	board := params.Get("board")
	var req *http.Request
	if params.Get("from") == "" {
		req, err = http.NewRequest("GET", NORMAL_API+"/"+board+"/get-all", nil)
	} else {
		offset, err := strconv.Atoi(params.Get("from"))
		if err != nil {
			http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
			return
		}
		to, err := strconv.Atoi(params.Get("to"))
		if err != nil {
			http.Error(w, ErrBadRequest.Error(), http.StatusInternalServerError)
			return
		}
		n := to - offset
		req, err = http.NewRequest("GET", fmt.Sprintf("%v/%v/get-%v-%v", NORMAL_API, board, offset, n), nil)
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

func getComments(w http.ResponseWriter, r *http.Request) {
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
	n := to - offset
	req, err = http.NewRequest("GET", fmt.Sprintf("%v/get-responses-%v-%v-%v", NORMAL_API, id, offset, n), nil)
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
