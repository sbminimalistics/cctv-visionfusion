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

// Create inserts a new event row into PostgreSQL.
func (r *DetectionRepo) Create(ctx context.Context, ev *model.CapturedEvent) error {
	return r.db.WithContext(ctx).Create(ev).Error
}

// GetByUUID fetches a single event by its UUIDv7 primary key.
func (r *DetectionRepo) GetByUUID(ctx context.Context, eventID uuid.UUID) (*model.CapturedEvent, error) {
	var ev model.CapturedEvent
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
