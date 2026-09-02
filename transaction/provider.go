// Package transaction allow running operations with multiple operations in a single transaction.
package transaction

import (
	"context"

	"gorm.io/gorm"
)

// Provider allows running multiple operations within a transaction, a custom function is used
// to construct new repositories or adapters of type T running within a transaction.
//
// Safe for concurrent use.
type Provider[T any] struct {
	db *gorm.DB

	// constructor function.
	newFn func(tx *gorm.DB) T
}

// NewProvider creates a new [Provider] with the given db connection and constructor function. The
// constructor function must construct required adapters that run its operations within the given
// transaction.
func NewProvider[T any](db *gorm.DB, newFn func(tx *gorm.DB) T) *Provider[T] {
	return &Provider[T]{
		db:    db,
		newFn: newFn,
	}
}

// Transact runs a closure within a transaction.
func (p *Provider[T]) Transact(ctx context.Context, runFn func(adapters T) error) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		adapters := p.newFn(tx)

		return runFn(adapters)
	})
}
