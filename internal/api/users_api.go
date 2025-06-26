package api

import (
	"board/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/schema"
)

var secret_key = []byte("да иди нахуй, ты вообще не должен был это читать")

const TOKEN_VALID_TIME = 24 //сколько валиден токен в часах
const MAX_NAME_LEN = 40     //максимальная длина имени

var ErrCantBeAnon = fmt.Errorf("нельзя быть аноном")
var ErrForbiddenChars = fmt.Errorf("уберите запятую")

func (a *Api) CreateUser(w http.ResponseWriter, r *http.Request) {
	if postRateLimiter.RespondOnLimit(w, r, r.RemoteAddr) {
		return
	}
	var user models.User
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
	if user.Name == ANON {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(ErrBadMan.Error()))
		return
	}
	user.Name = string([]rune(user.Name)[:min(MAX_NAME_LEN, len([]rune(user.Name)))])
	if strings.Contains(user.Name, ",") {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(ErrForbiddenChars.Error()))
		return
	}
	user.Admin = false
	t, err := bcrypt.GenerateFromPassword([]byte(user.Pass), 10)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	user.Pass = string(t)
	err = a.Repo.CreateUser(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("Ok"))
}

var ErrWrongPassword = fmt.Errorf("неправославный пароль или нечестивое имя")

func (a *Api) Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
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
	userInBd, err := a.Repo.GetUser(user.Name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(userInBd.Pass), []byte(user.Pass)) != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(ErrWrongPassword.Error()))
		return
	}
	claims := jwt.MapClaims{
		"name": user.Name,
		"exp":  time.Now().Add(time.Hour * TOKEN_VALID_TIME).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret_key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	resultJson, err := json.Marshal(map[string]string{"JWT": tokenString})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(resultJson)
}

func (a *Api) Op(w http.ResponseWriter, r *http.Request) {
	var user models.User
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
	userInCtx, ok := r.Context().Value("user").(models.User)
	if !ok || !userInCtx.Admin {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(ErrNotAdmin.Error()))
		return
	}
	a.Repo.UpdateUserAdmin(user.Name, true)
	w.Write([]byte("OK"))
}

func parseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret_key, nil
	})
}

var ErrInvalidToken = fmt.Errorf("ваш токен протух, перелогиньтесь")

func (a *Api) ValidateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		//сначала куки
		cookie, err := r.Cookie("JWT")
		if err != nil {
			tokenString = ""
		} else {
			tokenString = cookie.Value
		}
		//потом голова, если что головой перепишем
		h := r.Header.Get("JWT")
		if h != "" {
			tokenString = h
		}
		if tokenString != "" {
			token, err := parseToken(tokenString)
			if err != nil || !token.Valid {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(ErrInvalidToken.Error()))
				return
				//next.ServeHTTP(w, r)
			} else {
				claims := token.Claims.(jwt.MapClaims)
				name, ok := claims["name"].(string)
				if !ok {
					next.ServeHTTP(w, r)
				} else {
					user, err := a.Repo.GetUser(name)
					if err != nil {
						next.ServeHTTP(w, r)
					} else {
						ctx := context.WithValue(r.Context(), "user", user)
						next.ServeHTTP(w, r.WithContext(ctx))
					}
				}
			}
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
