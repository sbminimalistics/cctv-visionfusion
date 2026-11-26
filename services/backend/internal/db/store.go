package db

import (
	"context"

	"gorm.io/gorm"
)

type txCtxKey struct{}

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

func (s *Store) WithTx(ctx context.Context, fn func(txCtx context.Context, txStore *Store) error) error {
	// reuse existing transaction if already inside an active WithTx block
	if txDB, ok := ctx.Value(txCtxKey{}).(*gorm.DB); ok && txDB != nil {
		return fn(ctx, NewStore(txDB.WithContext(ctx)))
	}

	// start a new transaction and store it in txCtx for any nested calls
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txCtxKey{}, tx)
		return fn(txCtx, NewStore(tx.WithContext(txCtx)))
	})
}

// ExtractTx returns the active transactional *gorm.DB from ctx if one exists,
// otherwise falls back to fallbackDB.WithContext(ctx).
func ExtractTx(ctx context.Context, fallbackDB *gorm.DB) *gorm.DB {
	if txDB, ok := ctx.Value(txCtxKey{}).(*gorm.DB); ok && txDB != nil {
		return txDB.WithContext(ctx)
	}
	return fallbackDB.WithContext(ctx)
}
