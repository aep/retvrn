package graph

import (
	"context"
	"github.com/google/uuid"
	"iter"
	"retvrn/kv"
)

// Returns an iterator over all edges originating from the id and key
func Get(r kv.Read, ctx context.Context, id uuid.UUID, key string) iter.Seq2[uuid.UUID, error] {
	return func(yield func(uuid.UUID, error) bool) {
	}
}

// Set a key associated with a specific id, to a value
func Set(w kv.Write, from uuid.UUID, key string, to uuid.UUID) error {
	return nil
}
