package graph

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"iter"
	"retvrn/kv"
)

func GetN(r kv.Read, ctx context.Context, id uuid.UUID, key string) iter.Seq2[uuid.UUID, error] {
	if err := checkValidKey(key); err != nil {
		return func(yield func(uuid.UUID, error) bool) {
			yield(uuid.UUID{}, err)
		}
	}

	start := []byte(fmt.Sprintf("s.%s.%s.", key, id))
	end := bytes.Clone(start)
	end[len(end)-1] += 1

	return func(yield func(uuid.UUID, error) bool) {

		var visited = map[uuid.UUID]bool{}

		for kv, err := range r.Iter(ctx, start, end) {

			if err != nil {
				if yield(uuid.UUID{}, err) {
					continue
				}
				return
			}
			splitHere := bytes.LastIndexByte(kv.K, '.')

			id, err := uuid.Parse(string(kv.K[splitHere+1:]))
			if err != nil {
				continue
			}

			if visited[id] {
				continue
			}
			visited[id] = true

			if !yield(id, nil) {
				return
			}
		}
	}
}

func Get1(r kv.Read, ctx context.Context, id uuid.UUID, key string) (uuid.UUID, error) {

	k := []byte(fmt.Sprintf("s.%s.%s.unique", key, id))

	b, err := r.Get(ctx, k)
	if err != nil {
		return uuid.UUID{}, err
	}

	u, err := uuid.Parse(string(b))
	return u, err

}

func PutN(w kv.Write, from uuid.UUID, key string, to uuid.UUID) error {
	if err := checkValidKey(key); err != nil {
		return err
	}

	k := []byte(fmt.Sprintf("s.%s.%s.%s", key, from, to))

	return w.Put(k, []byte{0})
}

func Put1(w kv.Write, from uuid.UUID, key string, to uuid.UUID) error {
	if err := checkValidKey(key); err != nil {
		return err
	}

	k := []byte(fmt.Sprintf("s.%s.%s.unique", key, from))

	return w.Put(k, []byte(to.String()))
}

func Del1(w kv.Write, from uuid.UUID, key string) error {
	if err := checkValidKey(key); err != nil {
		return err
	}

	k := []byte(fmt.Sprintf("s.%s.%s.unique", key, from))

	return w.Del(k)
}

func DelN(w kv.Write, from uuid.UUID, key string, to uuid.UUID) error {
	if err := checkValidKey(key); err != nil {
		return err
	}

	k := []byte(fmt.Sprintf("s.%s.%s.%s", key, from, to))

	return w.Del(k)
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
