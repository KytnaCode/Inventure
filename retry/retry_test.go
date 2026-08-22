package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kytnacode/inventure/retry"
)

var errRealError = errors.New("real error")

func TestDoConfigShouldSuccessOnAlwaysSuccessAction(t *testing.T) {
	t.Parallel()

	action := func(_ context.Context) (bool, error) {
		return false, nil
	}

	conf := &retry.Config{
		MaxAttempts:    5,
		InitialBackoff: time.Millisecond,
		JitterRange:    time.Millisecond,
		BackoffFactor:  2,
	}

	err := retry.DoConfig(t.Context(), conf, action)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoConfigShouldSuccessOnTemporaryFailureAction(t *testing.T) {
	t.Parallel()

	once := false

	action := func(_ context.Context) (bool, error) {
		if !once {
			once = true

			return true, errRealError
		}

		return false, nil
	}

	conf := &retry.Config{
		MaxAttempts:    2,
		InitialBackoff: time.Millisecond,
		JitterRange:    time.Millisecond,
		BackoffFactor:  2,
	}

	err := retry.DoConfig(t.Context(), conf, action)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoConfigShouldFailInmediatelyOnNotTemporalErrors(t *testing.T) {
	t.Parallel()

	once := false

	action := func(_ context.Context) (bool, error) {
		if !once {
			once = true

			return false, errRealError
		}

		return false, nil
	}

	conf := &retry.Config{
		MaxAttempts:    2,
		InitialBackoff: time.Millisecond,
		JitterRange:    time.Millisecond,
		BackoffFactor:  2,
	}

	err := retry.DoConfig(t.Context(), conf, action)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDoConfigShouldRetryOnTemporaryError(t *testing.T) {
	t.Parallel()

	const maxAttempts = 5

	var count int

	action := func(_ context.Context) (bool, error) {
		if count < maxAttempts-1 {
			count++

			return true, errRealError
		}

		return false, nil
	}

	conf := &retry.Config{
		MaxAttempts:    maxAttempts,
		InitialBackoff: time.Millisecond,
		JitterRange:    time.Millisecond,
		BackoffFactor:  2,
	}

	err := retry.DoConfig(t.Context(), conf, action)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoConfigShouldCancelOnContextDone(t *testing.T) {
	t.Parallel()

	action := func(ctx context.Context) (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(time.Second * 5):
			return false, errRealError
		}
	}

	conf := &retry.Config{
		MaxAttempts:    3,
		InitialBackoff: time.Second,
		JitterRange:    time.Millisecond,
		BackoffFactor:  2,
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond)
	defer cancel()

	err := retry.DoConfig(ctx, conf, action)
	if err == nil {
		t.Error("expected an error")
	}

	if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected canceled error: got '%v'", err)
	}
}
