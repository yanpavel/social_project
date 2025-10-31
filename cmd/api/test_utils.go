package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yanpavel/social_project/internal/auth"
	"github.com/yanpavel/social_project/internal/store"
	"github.com/yanpavel/social_project/internal/store/cache"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T, cfg config) *application {
	t.Helper()

	logger := zap.Must(zap.NewProduction()).Sugar()
	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewMockStore()

	testAuth := &auth.TestAuthenticator{}

	return &application{
		logger:        logger,
		store:         mockStore,
		cacheStorage:  mockCacheStore,
		authenticator: testAuth,
		config:        cfg,
	}
}

func execRequest(mux http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d", expected, actual)
	}
}
