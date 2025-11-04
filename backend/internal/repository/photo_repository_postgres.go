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

// PhotoRepositoryPostgres implements PhotoRepository using PostgreSQL
type PhotoRepositoryPostgres struct {
	db     *database.PostgresDB
	logger *zap.Logger
}

// NewPhotoRepositoryPostgres creates a new PostgreSQL photo repository
func NewPhotoRepositoryPostgres(db *database.PostgresDB, logger *zap.Logger) domain.PhotoRepository {
	return &PhotoRepositoryPostgres{
		db:     db,
		logger: logger,
	}
}

// Create creates a new photo
func (r *PhotoRepositoryPostgres) Create(ctx context.Context, photo *domain.Photo) error {
	query := `
		INSERT INTO photos (id, user_id, url, thumbnail_url, caption, location, 
			taken_at, uploaded_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.GetDB().ExecContext(ctx, query,
		photo.ID,
		photo.UserID,
		photo.FilePath,
		photo.FileSize,
		photo.MimeType,
		photo.Description,
		photo.Location,
		photo.TakenAt,
		photo.UploadedBy,
		photo.CreatedAt,
		photo.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create photo", zap.Error(err))
		return err
	}

	return nil
}

// FindByID finds a photo by ID
func (r *PhotoRepositoryPostgres) FindByID(ctx context.Context, id string) (*domain.Photo, error) {
	query := `
		SELECT id, user_id, url, thumbnail_url, caption, location,
			taken_at, uploaded_by, created_at, updated_at
		FROM photos WHERE id = $1
	`

	photo := &domain.Photo{}
	var takenAt sql.NullTime

	err := r.db.GetDB().QueryRowContext(ctx, query, id).Scan(
		&photo.ID,
		&photo.UserID,
		&photo.FilePath,
		&photo.FileSize,
		&photo.MimeType,
		&photo.Description,
		&photo.Location,
		&takenAt,
		&photo.UploadedBy,
		&photo.CreatedAt,
		&photo.UpdatedAt,
	)

	if takenAt.Valid {
		photo.TakenAt = &takenAt.Time
	}

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("photo not found")
	}
	if err != nil {
		r.logger.Error("Failed to find photo by ID", zap.Error(err))
		return nil, err
	}

	return photo, nil
}

// FindByUserID finds photos by user ID
func (r *PhotoRepositoryPostgres) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Photo, error) {
	query := `
		SELECT id, user_id, url, thumbnail_url, caption, location,
			taken_at, uploaded_by, created_at, updated_at
		FROM photos 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.GetDB().QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to find photos by user ID", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var photos []*domain.Photo
	for rows.Next() {
		photo := &domain.Photo{}
		var takenAt sql.NullTime

		err := rows.Scan(
			&photo.ID,
			&photo.UserID,
			&photo.FilePath,
			&photo.FileSize,
			&photo.MimeType,
			&photo.Description,
			&photo.Location,
			&takenAt,
			&photo.UploadedBy,
			&photo.CreatedAt,
			&photo.UpdatedAt,
		)

		if err != nil {
			r.logger.Error("Failed to scan photo", zap.Error(err))
			continue
		}

		if takenAt.Valid {
			photo.TakenAt = &takenAt.Time
		}

		photos = append(photos, photo)
	}

	return photos, nil
}

// Update updates a photo
func (r *PhotoRepositoryPostgres) Update(ctx context.Context, photo *domain.Photo) error {
	query := `
		UPDATE photos SET
			description = $2,
			location = $3,
			taken_at = $4,
			updated_at = $5
		WHERE id = $1
	`

	photo.UpdatedAt = time.Now()

	_, err := r.db.GetDB().ExecContext(ctx, query,
		photo.ID,
		photo.Description,
		photo.Location,
		photo.TakenAt,
		photo.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to update photo", zap.Error(err))
		return err
	}

	return nil
}

// Delete deletes a photo
func (r *PhotoRepositoryPostgres) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM photos WHERE id = $1`

	_, err := r.db.GetDB().ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete photo", zap.Error(err))
		return err
	}

	return nil
}

// CountByUserID counts photos by user ID
func (r *PhotoRepositoryPostgres) CountByUserID(ctx context.Context, userID string) (int64, error) {
	query := `SELECT COUNT(*) FROM photos WHERE user_id = $1`

	var count int64
	err := r.db.GetDB().QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to count photos", zap.Error(err))
		return 0, err
	}

	return count, nil
}
