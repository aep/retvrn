package kv

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"iter"
	"reflect"
	"strings"
)

type Object struct{}

// Set a value by key and id
func (i Object) Set(w Write, id uuid.UUID, key string, value interface{}) error {

	if err := checkValidKey(key); err != nil {
		return err
	}

	k := []byte(fmt.Sprintf("f.%s.%s", key, id))

	v, err := serialize(value)
	if err != nil {
		return err
	}

	w.Set(k, v)

	return nil
}

// Get a value by key and id.
func (i Object) Get(r Read, ctx context.Context, id uuid.UUID, key string, value interface{}) error {

	if err := checkValidKey(key); err != nil {
		return err
	}

	k := []byte(fmt.Sprintf("f.%s.%s", key, id))

	v, err := r.Get(ctx, k)
	if err != nil {
		return err
	}

	return deserialize(v, value)
}

type Search struct{}

// WARNING: whatever we do here must be predictable and we need to maintain it forever,
// because to delete a key we must be able to reproduce the exact terms
func (i Search) terms(value interface{}) ([][]byte, error) {

	var terms [][]byte

	// first search term is just the exact value
	v, err := serialize(value)
	if err != nil {
		return nil, err
	}

	if len(v) > 255 {
		terms = append(terms, v[:255])
	} else {
		terms = append(terms, v)
	}

	if reflect.TypeOf(value).Kind() == reflect.String {

		v := []byte("0s" + strings.ToLower(value.(string)))

		if len(v) > 255 {
			terms = append(terms, v[:255])
		} else {
			terms = append(terms, v)
		}

		// the simplest possible fulltext search is just splitting the text
		for _, word := range strings.Fields(value.(string)) {
			v := []byte("0s" + strings.ToLower(word))
			if len(v) > 255 {
				terms = append(terms, v[:255])
			} else {
				terms = append(terms, v)
			}
		}
	}

	return terms, nil
}

func (i Search) Set(w Write, id uuid.UUID, key string, value interface{}) error {
	if err := checkValidKey(key); err != nil {
		return err
	}

	terms, err := i.terms(value)
	if err != nil {
		return err
	}

	for _, term := range terms {
		k := append([]byte(fmt.Sprintf("s.%s.", key)), term...)
		k = append(k, 0)
		k = append(k, '.')
		k = append(k, []byte(id.String())...)
		w.Set(k, []byte{0})
	}

	return nil

}

func (i Search) Get(r Read, ctx context.Context, key string, value interface{}) iter.Seq2[uuid.UUID, error] {

	if err := checkValidKey(key); err != nil {
		return func(yield func(uuid.UUID, error) bool) {
			yield(uuid.UUID{}, err)
		}
	}

	terms := [][]byte{}

	// first find by exact value
	vv, err := serialize(value)
	if err != nil {
		return func(yield func(uuid.UUID, error) bool) {
			yield(uuid.UUID{}, err)
		}
	}
	terms = append(terms, vv)

	if reflect.TypeOf(value).Kind() == reflect.String {
		//lowercase
		terms = append(terms, []byte("0s"+strings.ToLower(value.(string))))
	}

	return func(yield func(uuid.UUID, error) bool) {

		var visited = map[uuid.UUID]bool{}

		for _, term := range terms {

			start := append([]byte(fmt.Sprintf("s.%s.", key)), term...)
			end := bytes.Clone(start)
			end[len(end)-1] += 1

			for kv, err := range r.Iter(ctx, start, end) {
				if err != nil {
					if yield(uuid.UUID{}, err) {
						continue
					}
					return
				}
				splitHere := bytes.LastIndexByte(kv.K, '.')
				if splitHere+1 >= len(kv.K) {
					continue
				}
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
}

type Graph struct{}

// Returns an iterator over all edges originating from the id and key
func (i Graph) Get(r Read, ctx context.Context, id uuid.UUID, key string) iter.Seq2[uuid.UUID, error] {
	return func(yield func(uuid.UUID, error) bool) {
	}
}

// Set a key associated with a specific id, to a value
func (i Graph) Set(w Write, from uuid.UUID, key string, to uuid.UUID) error {
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
