package repo

import (
	"board/internal/models"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// названия полей в бд
const (
	ID        = "id"
	AUTHOR    = "author"
	TEXT      = "postText"
	TIMESTAMP = "postTime"
	DATA      = "data"
	PARENT    = "parentId"
	BOARD     = "board"
)

// названия таблиц в бд
const (
	POSTS_TABLE = "posts"
)

const LOGFILE_NAME = "postsrepo.log"

// конфиг репозитория
type RepoConfig struct {
	UserName string `env:"POSTGRES_USER" env-default:"root"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"123"`
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	DbName   string `env:"POSTGRES_DB" env-default:"db"`

	MaxPostsSelect int `env:"MAX_POSTS" env-default:"100"`
}

type Repo struct {
	db     *sqlx.DB
	logger *log.Logger
	RepoConfig
}

func NewPostsRepo(c RepoConfig) (*Repo, error) {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable host=%s port=%s",
		c.UserName, c.Password, c.DbName, c.Host, c.Port)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	if _, err := db.Conn(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logfile, err := os.OpenFile(LOGFILE_NAME, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	logger := log.New(logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
	return &Repo{db: db, logger: logger, RepoConfig: c}, nil
}

func (pr *Repo) CreatePost(post models.Post) error {
	_, err := sq.Insert(POSTS_TABLE).
		Columns(AUTHOR, TEXT, TIMESTAMP, DATA, PARENT, BOARD).
		Values(post.Author, post.Text, post.Timestamp, post.Data, post.ParentId, post.Board).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db).Exec()
	return err
}

func (pr *Repo) GetNPostsFromBoard(n int, offset int, board string, reversed bool, all bool) ([]models.Post, error) {
	if n <= 0 {
		return nil, fmt.Errorf("n should be positive")
	}
	if offset < 0 {
		return nil, fmt.Errorf("offset should be non-negative")
	}
	t := ""
	if reversed {
		t = " DESC"
	} else {
		t = " ASC"
	}
	n = min(n, pr.MaxPostsSelect)
	var rows *sql.Rows
	var err error
	if all {
		rows, err = sq.Select(ID, AUTHOR, TEXT, TIMESTAMP, DATA, PARENT, BOARD).
			From(POSTS_TABLE).
			Where(sq.Eq{BOARD: board}).
			Offset(uint64(offset)).
			Limit(uint64(n)).
			OrderBy(ID + t).
			PlaceholderFormat(sq.Dollar).
			RunWith(pr.db).Query()
	} else {
		rows, err = sq.Select(ID, AUTHOR, TEXT, TIMESTAMP, DATA, PARENT, BOARD).
			From(POSTS_TABLE).
			Where(sq.Eq{PARENT: 0, BOARD: board}).
			Offset(uint64(offset)).
			Limit(uint64(n)).
			OrderBy(ID + t).
			PlaceholderFormat(sq.Dollar).
			RunWith(pr.db).Query()
	}
	if err != nil {
		return nil, err
	}
	posts := make([]models.Post, 0)
	for rows.Next() {
		var post models.Post
		rows.Scan(&post.Id, &post.Author, &post.Text, &post.Timestamp, &post.Data, &post.ParentId, &post.Board)
		posts = append(posts, post)
	}
	return posts, nil
}

func (pr *Repo) GetAllPostsIdFromBoard(board string, reversed bool, all bool) ([]int, error) {
	/*if n <= 0 {
		return nil, fmt.Errorf("n should be positive")
	}
	if offset < 0 {
		return nil, fmt.Errorf("offset should be non-negative")
	}*/
	t := ""
	if reversed {
		t = " DESC"
	} else {
		t = " ASC"
	}
	//n = min(n, pr.MaxPostsSelect)
	var rows *sql.Rows
	var err error
	if all {
		rows, err = sq.Select(ID).
			From(POSTS_TABLE).
			Where(sq.Eq{BOARD: board}).
			OrderBy(ID + t).
			PlaceholderFormat(sq.Dollar).
			RunWith(pr.db).Query()
	} else {
		rows, err = sq.Select(ID).
			From(POSTS_TABLE).
			Where(sq.Eq{PARENT: 0, BOARD: board}).
			OrderBy(ID + t).
			PlaceholderFormat(sq.Dollar).
			RunWith(pr.db).Query()
	}
	if err != nil {
		return nil, err
	}
	posts := make([]int, 0)
	for rows.Next() {
		var i int
		rows.Scan(&i)
		posts = append(posts, i)
	}
	return posts, nil
}

func (pr *Repo) GetResponsesForPost(op int, offset int, n int, reversed bool) ([]models.Post, error) {
	t := ""
	if reversed {
		t = " DESC"
	} else {
		t = " ASC"
	}
	n = min(n, pr.MaxPostsSelect)
	rows, err := sq.Select(ID, AUTHOR, TEXT, TIMESTAMP, DATA, PARENT, BOARD).
		From(POSTS_TABLE).
		Where(sq.Eq{PARENT: op}).
		Offset(uint64(offset)).
		Limit(uint64(n)).
		OrderBy(ID + t).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db).Query()
	if err != nil {
		return nil, err
	}
	posts := make([]models.Post, 0)
	for rows.Next() {
		var post models.Post
		rows.Scan(&post.Id, &post.Author, &post.Text, &post.Timestamp, &post.Data, &post.ParentId, &post.Board)
		posts = append(posts, post)
	}
	return posts, nil
}

func (pr *Repo) GetAllResponsesForPost(op int, reversed bool) ([]int, error) {
	t := ""
	if reversed {
		t = " DESC"
	} else {
		t = " ASC"
	}
	//n = min(n, pr.MaxPostsSelect)
	rows, err := sq.Select(ID).
		From(POSTS_TABLE).
		Where(sq.Eq{PARENT: op}).
		//Offset(uint64(offset)).
		//Limit(uint64(n)).
		OrderBy(ID + t).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db).Query()
	if err != nil {
		return nil, err
	}
	posts := make([]int, 0)
	for rows.Next() {
		var post int
		rows.Scan(&post)
		posts = append(posts, post)
	}
	return posts, nil
}

func (pr *Repo) GetPost(id int64) (models.Post, error) {
	post := models.Post{}
	err := sq.Select(ID, AUTHOR, TEXT, TIMESTAMP, DATA, PARENT, BOARD).
		From(POSTS_TABLE).
		Where(sq.Eq{ID: id}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db).QueryRow().Scan(
		&post.Id, &post.Author, &post.Text, &post.Timestamp, &post.Data, &post.ParentId, &post.Board)
	if err != nil {
		return models.Post{}, err
	}
	return post, nil
}

/*
func (ur *UserRepo) GetUserByID(id int) (models.User, error) {
	user := models.User{}
	err := sq.Select(
		ID, NAME, PHONE, STATE, CONFIRMATION_CODE, CODE_SENT_AT, IS_PHONE_CONFIRMED).
		From(USERS_TABLE).
		Where(sq.Eq{ID: id}).
		PlaceholderFormat(sq.Dollar).
		RunWith(ur.db).QueryRow().Scan(
		&user.ID, &user.Name, &user.Phone, &user.State, &user.ConfirmationCode,
		&user.CodeSentAt, &user.IsPhoneConfirmed,
	)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (ur *UserRepo) CreateUser(user models.User) error {
	_, err := sq.Insert(USERS_TABLE).
		Columns(ID, NAME, PHONE, STATE, CONFIRMATION_CODE, CODE_SENT_AT, IS_PHONE_CONFIRMED).
		Values(user.ID, user.Name, user.Phone, user.State, user.ConfirmationCode,
			user.CodeSentAt, user.IsPhoneConfirmed).
		PlaceholderFormat(sq.Dollar).
		RunWith(ur.db).Exec()
	return err
}

func (ur *UserRepo) UpdateUser(user models.User) error {
	_, err := sq.Update(USERS_TABLE).
		SetMap(map[string]interface{}{
			NAME:               user.Name,
			PHONE:              user.Phone,
			STATE:              user.State,
			CONFIRMATION_CODE:  user.ConfirmationCode,
			CODE_SENT_AT:       user.CodeSentAt,
			IS_PHONE_CONFIRMED: user.IsPhoneConfirmed,
		}).
		Where(sq.Eq{ID: user.ID}).
		PlaceholderFormat(sq.Dollar).
		RunWith(ur.db).Exec()
	return err
}
*/
