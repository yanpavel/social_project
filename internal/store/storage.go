package store

import (
	"context"
	"database/sql"
)

type Storage struct {
	Posts interface {
		GetPostById(context.Context, int64) (*Post, error)
		Create(context.Context, *Post) error
		UpdatePost(context.Context, *Post, int64) error
		DeletePost(context.Context, int64) error
	}
	Users interface {
		Create(context.Context, *User) error
	}
	Comments interface {
		GetByPostID(context.Context, int64) (*[]Comment, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostStore{db},
		Users:    &UserStore{db},
		Comments: &CommentsStore{db},
	}
}
