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
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrDuplicateUsername = errors.New("username already exists")
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
		Create(context.Context, *User, *sql.Tx) error
		GetUserById(context.Context, int64) (*User, error)
		CreateAndInvite(context.Context, *User, string, time.Duration) error
		CreateUserInvitation(context.Context, *sql.Tx, string, time.Duration, int64) error
		Activate(context.Context, string) error
		Delete(context.Context, int64) error
		GetByEmail(context.Context, string) (*User, error)
	}
	Comments interface {
		GetByPostID(context.Context, int64) (*[]Comment, error)
		CreateComment(context.Context, *Comment) (*int64, error)
		GetCommentById(context.Context, int64) (*Comment, error)
		DeleteCommentById(context.Context, int64) error
		UpdateCommentById(context.Context, *Comment) error
	}
	Followers interface {
		FollowUser(context.Context, int64, int64) error
		UnfollowUser(context.Context, int64, int64) error
	}
	Roles interface {
		GetByName(context.Context, string) (*Role, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db},
		Users:     &UserStore{db},
		Comments:  &CommentsStore{db},
		Followers: &FollowersStore{db},
		Roles:     &RoleStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
