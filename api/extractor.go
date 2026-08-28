package api

import "context"

// Extractor tries to extract a value of type `T` from a context and return and error on fail.
type Extractor[T any] func(ctx context.Context) (T, error)
