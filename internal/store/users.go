package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id        int64    `json:"id"`
	UserName  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
	RoleID    int64    `json:"role_id"`
	Role      Role     `json:"role"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

func (p *password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, username, email, password, created_at FROM users
		WHERE email = $1 AND is_active = true
	`

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.Id,
		&user.UserName,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) GetUserById(ctx context.Context, userId int64) (*User, error) {
	query := `
	SELECT u.username, u.email, u.id, u.password, u.created_at, r.* 
	FROM users u
	JOIN roles r ON u.role_id=r.id
	WHERE u.id=$1 AND u.is_active=true
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		userId,
	).Scan(
		&user.UserName,
		&user.Email,
		&user.Id,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Role.Id,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

func (s *UserStore) Create(ctx context.Context, user *User, tx *sql.Tx) error {
	query := `INSERT INTO users (username, password, email, role_id)
	VALUES($1, $2, $3, (SELECT id FROM roles WHERE name = $4)) RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	role := user.Role.Name
	if role == "" {
		role = "user"
	}

	err := tx.QueryRowContext(
		ctx,
		query,
		user.UserName,
		user.Password.hash,
		user.Email,
		role,
	).Scan(
		&user.Id,
		&user.CreatedAt,
	)

	if err != nil {
		switch err.Error() {
		case `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.Create(ctx, user, tx); err != nil {
			return err
		}

		if err := s.CreateUserInvitation(ctx, tx, token, exp, user.Id); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) CreateUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, id int64) error {
	query := `
		INSERT INTO user_invitations(token, user_id, expiry) VALUES($1, $2, $3)
	`

	_, err := tx.ExecContext(ctx, query, token, id, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}

		if err := s.deleteUserInvitations(ctx, tx, user.Id); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `
	DELETE FROM user_invitations WHERE user_id=$1
	`
	_, err := tx.ExecContext(ctx, query, id)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
	UPDATE users SET username=$1, email=$2, is_active=$3 WHERE id=$4
	`
	_, err := tx.ExecContext(ctx, query, user.UserName, user.Email, user.IsActive, user.Id)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
	SELECT u.id, u.username, u.email, u.created_at, u.is_active
	FROM users u
	JOIN user_invitations ui ON u.id=ui.user_id
	WHERE ui.token=$1 AND ui.expiry>$2
	`
	user := &User{}
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.Id,
		&user.UserName,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) Delete(ctx context.Context, id int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.delete(ctx, tx, id); err != nil {
			return err
		}

		if err := s.deleteUserInvitations(ctx, tx, id); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `
	DELETE FROM users WHERE id=$1
	`
	_, err := tx.ExecContext(ctx, query, id)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}
