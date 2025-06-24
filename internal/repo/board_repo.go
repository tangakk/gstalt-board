package repo

import (
	"board/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

const (
	DESC       = "description"
	ADMINS     = "admins"
	MODE       = "mode"
	USERS      = "usersList"
	OWNER      = "owner"
	POSTS      = "posts"
	POSTSWITHR = "postsWithR"
)

const (
	BOARDS_TABLE = "boards"
)

func (pr *Repo) CreateBoard(board models.Board) error {
	_, err := sq.Insert(BOARDS_TABLE).
		Columns(NAME, DESC, ADMINS, MODE, USERS, OWNER).
		Values(board.Name, board.Description, pq.Array(board.Admins), board.Mode, pq.Array(board.UsersList), board.Owner).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db).Exec()
	return err
}

func (pr *Repo) GetBoard(name string) (models.Board, error) {
	board := models.Board{}
	err := sq.Select(NAME, DESC, ADMINS, MODE, USERS, OWNER, POSTS, POSTSWITHR).
		From(BOARDS_TABLE).
		Where(sq.Eq{NAME: name}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db).QueryRow().Scan(
		&board.Name, &board.Description, (*pq.StringArray)(&board.Admins), &board.Mode, (*pq.StringArray)(&board.UsersList), &board.Owner,
		&board.PostsCount, &board.PostsCountWithResponses)
	if err != nil {
		return models.Board{}, err
	}
	return board, nil
}

func (pr *Repo) UpdateBoard(board models.Board) error {
	_, err := sq.Update(BOARDS_TABLE).
		Where(sq.Eq{NAME: board.Name}).
		SetMap(map[string]interface{}{
			NAME: board.Name, DESC: board.Description, ADMINS: pq.Array(board.Admins),
			MODE: board.Mode, USERS: pq.Array(board.UsersList)}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db).Exec()
	return err
}

func (pr *Repo) GetAllBoards() ([]models.Board, error) {
	rows, err := sq.Select(NAME, DESC, ADMINS, MODE, USERS, OWNER, POSTS, POSTSWITHR).
		From(BOARDS_TABLE).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db).Query()
	if err != nil {
		return nil, err
	}
	boards := make([]models.Board, 0)
	for rows.Next() {
		var board models.Board
		rows.Scan(&board.Name, &board.Description, (*pq.StringArray)(&board.Admins), &board.Mode, (*pq.StringArray)(&board.UsersList), &board.Owner,
			&board.PostsCount, &board.PostsCountWithResponses)
		boards = append(boards, board)
	}
	return boards, nil
}
