package db

import (
	"context"

	"gorm.io/gorm"
)

type Store struct {
	db               *gorm.DB
	Events           *EventRepo
	Detections       *DetectionRepo
	DetectionObjects *DetectionObjectsRepo
}

func NewStore(gdb *gorm.DB) *Store {
	return &Store{
		db:               gdb,
		Events:           &EventRepo{db: gdb},
		Detections:       &DetectionRepo{db: gdb},
		DetectionObjects: &DetectionObjectsRepo{db: gdb},
	}
}

func (s *Store) WithTx(ctx context.Context, fn func(txStore *Store) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewStore(tx))
	})
}
