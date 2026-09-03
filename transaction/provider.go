// Package transaction allow running operations with multiple operations in a single transaction.
package transaction

import (
	"context"
	"sync"

	"gorm.io/gorm"
)

// GormProvider provider allow running multiple repositories inside the same gorm transaction.
//
// Safe for concurrent use.
type GormProvider[T any] struct {
	db *gorm.DB

	// constructor function.
	newFn func(tx *gorm.DB) T
}

var _ Provider[any] = &TestProvider[any]{}

// TestProvider is a [Provider] meant for testing purposes.
//
// Safe for concurrent use.
type TestProvider[T any] struct {
	newFn func() T
	mut   sync.Mutex
}

// Provider allows running multiple operations within a transaction, a custom function is used
// to construct new repositories or adapters of type T running within a transaction.
//
// Safe for concurrent use.
type Provider[T any] interface {
	Transact(ctx context.Context, runFn func(adapters T) error) error
}

// NewGormProvider creates a new [Provider] with the given db connection and constructor function.
// The constructor function must construct required adapters that run its operations within the
// given transaction.
func NewGormProvider[T any](db *gorm.DB, newFn func(tx *gorm.DB) T) *GormProvider[T] {
	return &GormProvider[T]{
		db:    db,
		newFn: newFn,
	}
}

// NewTestProvider creates a new [TestProvider].
func NewTestProvider[T any](newFn func() T) *TestProvider[T] {
	return &TestProvider[T]{
		newFn: newFn,
	}
}

// Transact runs a closure within a transaction.
func (p *GormProvider[T]) Transact(ctx context.Context, runFn func(adapters T) error) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		adapters := p.newFn(tx)

		return runFn(adapters)
	})
}

// Transact runs `runFn` with the adapters gave to [NewTestProvider], don't run a real transaction.
func (p *TestProvider[T]) Transact(_ context.Context, runFn func(adapters T) error) error {
	p.mut.Lock()
	defer p.mut.Unlock()

	adapters := p.newFn()

	return runFn(adapters)
}
