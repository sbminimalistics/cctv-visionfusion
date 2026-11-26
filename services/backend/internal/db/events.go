package db

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"gorm.io/gorm"
	"sbminimalistics.com/cctv-visionfusion/internal/db/model"
)

var ErrEventNotFound = errors.New("event not found")

type EventRepo struct {
	db *gorm.DB
}

// Create inserts a new event row into PostgreSQL.
func (r *EventRepo) Create(ctx context.Context, ev *model.CapturedEvent) error {
	return r.db.WithContext(ctx).Create(ev).Error
}

// GetByUUID fetches a single event by its UUIDv7 primary key.
func (r *EventRepo) GetByUUID(ctx context.Context, eventID uuid.UUID) (*model.CapturedEvent, error) {
	var ev model.CapturedEvent
	err := r.db.WithContext(ctx).
		Where("id = ?", eventID).
		First(&ev).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrEventNotFound
	}
	if err != nil {
		return nil, err
	}

	return &ev, nil
}

// ListRecent returns the most recent events ordered newest-first.
func (r *EventRepo) ListRecent(ctx context.Context, limit int) ([]model.CapturedEvent, error) {
	if limit <= 0 {
		limit = 50
	}

	var events []model.CapturedEvent
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error

	return events, err
}

// UpdateTraceID attaches or updates the OpenTelemetry trace_id on an existing event.
func (r *EventRepo) UpdateTraceID(ctx context.Context, eventID uuid.UUID, traceID string) error {
	res := r.db.WithContext(ctx).
		Model(&model.CapturedEvent{}).
		Where("id = ?", eventID).
		Update("trace_id", traceID)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrEventNotFound
	}
	return nil
}

// UpdateStatus updates status on an existing event.
func (r *EventRepo) UpdateStatus(ctx context.Context, eventUUID uuid.UUID, status string) error {
	db := ExtractTx(ctx, r.db)
	return db.Model(&model.CapturedEvent{}).
		Where("id = ?", eventUUID).
		Update("status", status).Error
}
