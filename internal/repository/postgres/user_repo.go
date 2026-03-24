package postgres

import (
	"calendar/internal/domain"
	"calendar/internal/domain/user"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (p *Postgres) GetUser(ctx context.Context, login string) (*user.User, error) {
	query := `
		SELECT id, login, password, created_at, email, telegram
		FROM users
		WHERE login = $1
	`
	return p.getUser(ctx, query, login)
}

func (p *Postgres) GetUserByUUID(ctx context.Context, id string) (*user.User, error) {
	query := `
		SELECT id, login, password, created_at, email, telegram
		FROM users
		WHERE id = $1
	`
	return p.getUser(ctx, query, id)
}

func (p *Postgres) getUser(
	ctx context.Context,
	query string,
	arg any,
) (*user.User, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, p.cfg.DB.Postgres.ReadTimeout)
	defer cancel()

	var u user.User
	err := p.db.QueryRow(
		ctxWithTimeout,
		query,
		arg,
	).Scan(
		&u.ID,
		&u.Login,
		&u.Password,
		&u.CreatedAt,
		&u.Email,
		&u.Telegram,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &u, nil
}

func (p *Postgres) SaveUser(ctx context.Context, u *user.User) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, p.cfg.DB.Postgres.WriteTimeout)
	defer cancel()

	query := `
		INSERT INTO users (id, login, password, created_at, email, telegram)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := p.db.Exec(
		ctxWithTimeout,
		query,
		u.ID,
		u.Login,
		u.Password,
		u.CreatedAt,
		u.Email,
		u.Telegram,
	)

	if err != nil {
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}
