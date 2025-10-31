package store

import (
	"context"
	"database/sql"
)

type Role struct {
	Id          int64  `json:"id"`
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoleStore struct {
	db *sql.DB
}

func (s *RoleStore) GetByName(ctx context.Context, roleName string) (*Role, error) {
	query := `SELECT id, name, description, level FROM roles WHERE name = $1`

	role := &Role{}
	err := s.db.QueryRowContext(ctx, query, roleName).Scan(&role.Id, &role.Name, &role.Description, &role.Level)
	if err != nil {
		return nil, err
	}

	return role, nil
}
