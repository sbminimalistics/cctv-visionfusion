package db

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"gorm.io/gorm"
	"sbminimalistics.com/cctv-visionfusion/internal/db/model"
)

var ErrDetectionNotFound = errors.New("detection not found")

type DetectionRepo struct {
	db *gorm.DB
}

// Create inserts a new detection row into PostgreSQL.
func (r *DetectionRepo) Create(ctx context.Context, d *model.Detection) error {
	return r.db.WithContext(ctx).Create(d).Error
}

// GetByID fetches a single detection by its ID primary key.
func (r *DetectionRepo) GetByID(ctx context.Context, eventID uuid.UUID) (*model.Detection, error) {
	var ev model.Detection
	err := r.db.WithContext(ctx).
		Where("id = ?", eventID).
		First(&ev).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDetectionNotFound
	}
	if err != nil {
		return nil, err
	}

	return &ev, nil
}
