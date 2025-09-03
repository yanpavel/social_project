package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type Follower struct {
	UserID string `json:"user_id"`
}

type FollowersStore struct {
	db *sql.DB
}

func (s *FollowersStore) FollowUser(ctx context.Context, followeeId int64, followerId int64) error {
	query := `
	INSERT INTO followers
	(user_id, follower_id)
	VALUES ($1, $2)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		followeeId,
		followerId,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}

	return nil
}

func (s *FollowersStore) UnfollowUser(ctx context.Context, followeeId int64, followerId int64) error {
	query := `
	DELETE FROM followers
	WHERE user_id=$1 AND follower_id=$2
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	result, err := s.db.ExecContext(
		ctx,
		query,
		followeeId,
		followerId,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("count rows affectted error: %v", err)
	}

	return nil
}
