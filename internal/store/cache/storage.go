package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/yanpavel/social_project/internal/store"
)

type Storage struct {
	Users interface {
		Get(context.Context, int64) (*store.User, error)
		Set(context.Context, *store.User) error
	}
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		Users: &UserStore{rdb: rdb},
	}
}
