package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yanpavel/social_project/internal/store"
)

type UserStore struct {
	rdb *redis.Client
}

const UserExpTime = time.Minute

func (s *UserStore) Get(ctx context.Context, id int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", id)
	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user store.User
	if data != "" {
		err = json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}

func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%v", user.Id)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.rdb.SetEx(ctx, cacheKey, json, UserExpTime).Err()
}
