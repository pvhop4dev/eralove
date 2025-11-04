package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/infrastructure/database"
	"go.uber.org/zap"
)

// UserRepositoryPostgres implements UserRepository using PostgreSQL
type UserRepositoryPostgres struct {
	db     *database.PostgresDB
	logger *zap.Logger
}

// NewUserRepositoryPostgres creates a new PostgreSQL user repository
func NewUserRepositoryPostgres(db *database.PostgresDB, logger *zap.Logger) domain.UserRepository {
	return &UserRepositoryPostgres{
		db:     db,
		logger: logger,
	}
}

// Create creates a new user
func (r *UserRepositoryPostgres) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, date_of_birth, gender, bio, avatar_url, 
			partner_id, is_email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.GetDB().ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.DateOfBirth,
		user.Gender,
		user.Bio,
		user.AvatarURL,
		user.PartnerID,
		user.IsEmailVerified,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create user", zap.Error(err))
		return err
	}

	return nil
}

// FindByID finds a user by ID
func (r *UserRepositoryPostgres) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, date_of_birth, gender, bio, avatar_url,
			partner_id, is_email_verified, created_at, updated_at
		FROM users WHERE id = $1
	`

	user := &domain.User{}
	err := r.db.GetDB().QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.DateOfBirth,
		&user.Gender,
		&user.Bio,
		&user.AvatarURL,
		&user.PartnerID,
		&user.IsEmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		r.logger.Error("Failed to find user by ID", zap.Error(err))
		return nil, err
	}

	return user, nil
}

// FindByEmail finds a user by email
func (r *UserRepositoryPostgres) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, date_of_birth, gender, bio, avatar_url,
			partner_id, is_email_verified, created_at, updated_at
		FROM users WHERE email = $1
	`

	user := &domain.User{}
	err := r.db.GetDB().QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.DateOfBirth,
		&user.Gender,
		&user.Bio,
		&user.AvatarURL,
		&user.PartnerID,
		&user.IsEmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		r.logger.Error("Failed to find user by email", zap.Error(err))
		return nil, err
	}

	return user, nil
}

// Update updates a user
func (r *UserRepositoryPostgres) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET
			email = $2,
			password_hash = $3,
			name = $4,
			date_of_birth = $5,
			gender = $6,
			bio = $7,
			avatar_url = $8,
			partner_id = $9,
			is_email_verified = $10,
			updated_at = $11
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.GetDB().ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.DateOfBirth,
		user.Gender,
		user.Bio,
		user.AvatarURL,
		user.PartnerID,
		user.IsEmailVerified,
		user.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to update user", zap.Error(err))
		return err
	}

	return nil
}

// Delete deletes a user
func (r *UserRepositoryPostgres) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.GetDB().ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete user", zap.Error(err))
		return err
	}

	return nil
}

// ExistsByEmail checks if a user exists by email
func (r *UserRepositoryPostgres) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.GetDB().QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check user existence", zap.Error(err))
		return false, err
	}

	return exists, nil
}
