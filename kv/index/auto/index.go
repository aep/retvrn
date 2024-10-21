package auto

import (
	"context"
	"github.com/google/uuid"
	"retvrn/kv"
	"retvrn/kv/index/column"
	"retvrn/kv/index/search"
)

func Put(w kv.Write, id uuid.UUID, key string, value interface{}) error {
	err := column.Put(w, id, key, value)
	if err != nil {
		return err
	}
	err = search.Put(w, id, key, value)
	if err != nil {
		return err
	}
	return nil
}

func Del(ctx context.Context, w kv.Write, id uuid.UUID, key string) error {

	val, err := column.Get(w, ctx, id, key, nil)
	if err != nil {
		return err
	}

	err = column.Del(w, id, key)
	if err != nil {
		return err
	}

	err = search.Del(w, id, key, val)
	if err != nil {
		return err
	}
	return nil
}
