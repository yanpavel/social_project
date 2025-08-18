package store

import (
	"context"
	"database/sql"
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

func (s *CommentsStore) GetByPostID(ctx context.Context, postID int64) (*[]Comment, error) {
	query := `
	SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, u.username, u.id 
	FROM comments c
	JOIN users u on u.id = c.user_id
	WHERE c.post_id=$1
	ORDER BY c.created_at DESC
	`

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
			return nil, err
		}
	}

	return &comments, nil
}
