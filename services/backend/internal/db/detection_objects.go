package db

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"gorm.io/gorm"
	"sbminimalistics.com/cctv-visionfusion/internal/db/model"
)

var ErrDetectionObjectsNotFound = errors.New("detection objects not found")

type DetectionObjectsRepo struct {
	db *gorm.DB
}

// Create inserts a new detection object row into PostgreSQL.
func (r *DetectionObjectsRepo) Create(ctx context.Context, obj *model.DetectionObject) error {
	return r.db.WithContext(ctx).Create(obj).Error
}

// CreateBatch inserts multiple detection object rows into PostgreSQL.
func (r *DetectionObjectsRepo) CreateBatch(ctx context.Context, objs []*model.DetectionObject) error {
	if len(objs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(objs).Error
}

// GetByUUID fetches a single detection object by its ID primary key.
func (r *DetectionObjectsRepo) GetByID(ctx context.Context, eventID uuid.UUID) (*model.DetectionObject, error) {
	var ev model.DetectionObject
	err := r.db.WithContext(ctx).
		Where("id = ?", eventID).
		First(&ev).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDetectionObjectsNotFound
	}
	if err != nil {
		return nil, err
	}

	return &ev, nil
}
