package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/yanpavel/social_project/internal/store/cache"
)

func TestGetUser(t *testing.T) {
	withRedis := config{
		redisCfg: redisConfig{
			enabled: true,
		},
	}
	app := newTestApplication(t, withRedis)
	mux := app.mount()

	testToken, err := app.authenticator.GenerateToken(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("should not allow unauth request", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := execRequest(mux, req)

		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should allow auth requests", func(t *testing.T) {
		mockCacheStore := app.cacheStorage.Users.(*cache.MockUserStore)

		mockCacheStore.On("Get", int64(1)).Return(nil, nil).Twice()
		mockCacheStore.On("Set", mock.Anything).Return(nil)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Add("Authorization", "Bearer "+testToken)

		rr := execRequest(mux, req)

		checkResponseCode(t, http.StatusOK, rr.Code)

		mockCacheStore.Calls = nil
	})
	t.Run("should hit the cache first and if not exists it sets the user on the cache", func(t *testing.T) {
		mockCacheStore := app.cacheStorage.Users.(*cache.MockUserStore)

		c1 := mockCacheStore.On("Get", int64(42)).Return(nil, nil)
		c2 := mockCacheStore.On("Get", int64(1)).Return(nil, nil)
		c3 := mockCacheStore.On("Set", mock.Anything, mock.Anything).Return(nil)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Add("Authorization", "Bearer "+testToken)

		rr := execRequest(mux, req)

		checkResponseCode(t, http.StatusOK, rr.Code)

		fmt.Print(c1, c2, c3)
		mockCacheStore.AssertNumberOfCalls(t, "Get", 2)
		mockCacheStore.Calls = nil
	})
}
