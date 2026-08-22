// Package retry provides configurable logic for retrying actions with backoff and jitter.
package retry

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

// Default configuration parameters.
const (
	DefaultMaxAttempts    = 5
	DefaultInitialBackoff = time.Second * 5
	DefaultJitterRange    = time.Second
	DefaultBackoffFactor  = 2
	DefaultTimeout        = time.Minute
)

// Action defines a function that can be retried. The returned boolean indicates if the
// error is temporary (true) and should be retried, or permanent (false).
type Action func(context.Context) (temporary bool, err error)

// Config holds parameters for retry behavior. Use DefaultConfig() for sensible defaults.
type Config struct {
	// MaxAttempts is the maximum number of attempts before giving up. If a permanent error
	// occurs, retries stop immediately.
	MaxAttempts int

	// InitialBackoff is the initial wait time before retrying. It is multiplied by BackoffFactor
	// after each attempt.
	InitialBackoff time.Duration

	// JitterRange is the maximum random duration added to backoff to avoid synchronized
	// retries. Actual jitter is [0, JitterRange).
	JitterRange time.Duration

	// BackoffFactor multiplies the backoff after each failed attempt.
	BackoffFactor float64
}

// DefaultConfig returns a [Config] with sensible default values for retrying actions.
func DefaultConfig() *Config {
	return &Config{
		MaxAttempts:    DefaultMaxAttempts,
		InitialBackoff: DefaultInitialBackoff,
		JitterRange:    DefaultJitterRange,
		BackoffFactor:  DefaultBackoffFactor,
	}
}

// DoConfig retries the given action according to the provided [Config].
//
// If the action returns a temporary error, DoConfig retries up to MaxAttempts times,
// accumulating errors.
// If a permanent error occurs, it is returned immediately.
// If all attempts fail with temporary errors, a wrapped error containing all errors is returned.
// If the context is canceled, DoConfig returns a context cancellation error and stops retrying.
func DoConfig(ctx context.Context, conf *Config, act Action) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	backoff := conf.InitialBackoff

	var errs error

	for range conf.MaxAttempts {
		temp, err := act(ctx)
		if err != nil {
			if temp {
				errs = errors.Join(errs, err)

				//nolint:gosec // standard RNG is fine here.
				jitter := float64(conf.JitterRange) * rand.Float64()

				select {
				case <-ctx.Done():
					return fmt.Errorf("context canceled: %w", err)
				case <-time.After(backoff + time.Duration(jitter)):
					backoff = time.Duration(float64(backoff) * conf.BackoffFactor)

					continue
				}
			}

			return err
		}

		return nil
	}

	return errs
}

// Do retries the given action using the default retry configuration.
func Do(ctx context.Context, act Action) error {
	conf := DefaultConfig()

	return DoConfig(ctx, conf, act)
}
