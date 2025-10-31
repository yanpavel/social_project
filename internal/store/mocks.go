package store

import (
	"context"
	"database/sql"
	"time"
)

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct{}

func (m *MockUserStore) Create(ctx context.Context, u *User, tx *sql.Tx) error {
	return nil
}

func (m *MockUserStore) GetUserById(ctx context.Context, userID int64) (*User, error) {
	return &User{Id: userID}, nil
}

func (m *MockUserStore) GetByEmail(context.Context, string) (*User, error) {
	return &User{}, nil
}

func (m *MockUserStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return nil
}

func (m *MockUserStore) Activate(ctx context.Context, t string) error {
	return nil
}

func (m *MockUserStore) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *MockUserStore) CreateUserInvitation(context.Context, *sql.Tx, string, time.Duration, int64) error {
	return nil
}
