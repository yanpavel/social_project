package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exists")
	QueryTimeOutDuration = time.Second * 5
)

type Storage struct {
	Posts interface {
		GetPostById(context.Context, int64) (*Post, error)
		Create(context.Context, *Post) error
		UpdatePost(context.Context, *Post, int64) error
		DeletePost(context.Context, int64) error
		GetUserFeed(context.Context, int64, PaginationFeedQuery) ([]PostWithMetaData, error)
	}
	Users interface {
		Create(context.Context, *User) error
		GetUserById(context.Context, int64) (*User, error)
	}
	Comments interface {
		GetByPostID(context.Context, int64) (*[]Comment, error)
		CreateComment(context.Context, *Comment) (*int64, error)
	}
	Followers interface {
		FollowUser(context.Context, int64, int64) error
		UnfollowUser(context.Context, int64, int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db},
		Users:     &UserStore{db},
		Comments:  &CommentsStore{db},
		Followers: &FollowersStore{db},
	}
}
