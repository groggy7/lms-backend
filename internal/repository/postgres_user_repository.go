package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/serhatkilbas/lms-poc/internal/domain"
	_ "github.com/lib/pq"
)

type postgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) domain.UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(ctx context.Context, user domain.User) error {
	query := `INSERT INTO users (id, email, password, full_name, role) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.Password, user.FullName, string(user.Role))
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}
	return nil
}

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password, full_name, role FROM users WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, email)

	var user domain.User
	var role string
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FullName, &role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	user.Role = domain.Role(role)
	return &user, nil
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, password, full_name, role FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var user domain.User
	var role string
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FullName, &role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by id: %v", err)
	}
	user.Role = domain.Role(role)
	return &user, nil
}
