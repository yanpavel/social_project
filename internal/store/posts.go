package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

var (
	ErrNotFound = errors.New("resource not found")
)

type Post struct {
	Id        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserId    int64     `json:"user_id"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Tags      []string  `json:"tags"`
	Comments  []Comment `json:"comments"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `INSERT INTO posts (content, title, user_id, tags)
	VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserId,
		pq.Array(post.Tags),
	).Scan(
		&post.Id,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return err
	}
	return nil
}

func (s *PostStore) UpdatePost(ctx context.Context, post *Post, id int64) error {
	query := "UPDATE posts SET title=$1, content=$2, tags=$3 WHERE id=$4"

	result, err := s.db.ExecContext(
		ctx,
		query,
		post.Title,
		post.Content,
		pq.Array(post.Tags),
		post.Id,
	)

	if err != nil {
		return err
	}

	amount, err := result.RowsAffected()

	if amount == 0 {
		return ErrNotFound
	}

	if err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetPostById(ctx context.Context, id int64) (*Post, error) {
	query := `
	SELECT id, user_id, title, content, created_at, updated_at, tags
	FROM posts
	WHERE id=$1
	`
	var post Post
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.Id,
		&post.UserId,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags),
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound

		default:
			return nil, err
		}
	}

	return &post, nil
}

func (s *PostStore) DeletePost(ctx context.Context, id int64) error {
	query := "DELETE FROM posts WHERE id=$1"

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	amount, err := result.RowsAffected()

	if amount == 0 {
		return ErrNotFound
	}

	if err != nil {
		return err
	}

	return nil
}
