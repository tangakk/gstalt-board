package repo

import (
	"board/internal/models"

	sq "github.com/Masterminds/squirrel"
)

const (
	NAME  = "name"
	PASS  = "passHash"
	ADMIN = "admin"
)

const (
	USERS_TABLE = "users"
)

func (pr *Repo) CreateUser(user models.User) error {
	_, err := sq.Insert(USERS_TABLE).
		Columns(NAME, PASS, ADMIN).
		Values(user.Name, user.Pass, user.Admin).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db).Exec()
	return err
}

func (pr *Repo) GetUser(name string) (models.User, error) {
	user := models.User{}
	err := sq.Select(NAME, PASS, ADMIN).
		From(USERS_TABLE).
		Where(sq.Eq{NAME: name}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db).QueryRow().Scan(
		&user.Name, &user.Pass, &user.Admin)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (pr *Repo) UpdateUserAdmin(name string, admin bool) error {
	_, err := sq.Update(USERS_TABLE).
		Set(ADMIN, admin).Where(sq.Eq{NAME: name}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db).Exec()
	return err
}
