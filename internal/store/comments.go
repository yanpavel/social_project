package store

import (
	"context"
	"database/sql"
	"errors"
)

type Comment struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      User   `json: "user"`
}

type CommentsStore struct {
	db *sql.DB
}

func (s *CommentsStore) CreateComment(ctx context.Context, comment *Comment) (*int64, error) {
	query := `
	INSERT INTO comments
	(post_id, user_id, content)
	VALUES ($1, $2, $3)
	RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.PostID,
		comment.UserID,
		comment.Content,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	id := &comment.ID
	return id, nil
}

func (s *CommentsStore) GetByPostID(ctx context.Context, postID int64) (*[]Comment, error) {
	query := `
	SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, u.username, u.id 
	FROM comments c
	JOIN users u on u.id = c.user_id
	WHERE c.post_id=$1
	ORDER BY c.created_at DESC
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := []Comment{}
	for rows.Next() {
		var c Comment
		c.User = User{}
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.CreatedAt, &c.User.UserName, &c.User.Id); err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return nil, ErrNotFound
			default:
				return nil, err
			}
		}
		comments = append(comments, c)
	}

	return &comments, nil
}
