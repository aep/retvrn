package column

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"retvrn/kv"
	"retvrn/kv/index"
)

func Set(w kv.Write, id uuid.UUID, key string, value interface{}) error {

	if err := checkValidKey(key); err != nil {
		return err
	}

	k := []byte(fmt.Sprintf("f.%s.%s", key, id))

	v, err := index.Serialize(value)
	if err != nil {
		return err
	}

	w.Set(k, v)

	return nil
}

func Get(r kv.Read, ctx context.Context, id uuid.UUID, key string, value interface{}) (interface{}, error) {

	if err := checkValidKey(key); err != nil {
		return nil, err
	}

	k := []byte(fmt.Sprintf("f.%s.%s", key, id))

	v, err := r.Get(ctx, k)
	if err != nil {
		return nil, err
	}

	return index.Deserialize(v, value)
}

func Del(w kv.Write, id uuid.UUID, key string) error {

	if err := checkValidKey(key); err != nil {
		return err
	}

	k := []byte(fmt.Sprintf("f.%s.%s", key, id))

	w.Del(k)

	return nil
}

func checkValidKey(key string) error {
	if len(key) >= 255 {
		return fmt.Errorf("key must be < 255")
	}

	for _, c := range key {
		if c < '0' || c > 'z' {
			return fmt.Errorf("key must be alpha numeric")
		}
	}

	return nil
}
