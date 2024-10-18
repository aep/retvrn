package kv

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIndex(t *testing.T) {

	// open tikv

	var ctx = context.Background()
	k, err := NewTikv()
	require.NoError(t, err)
	defer k.Close()

	var object = Object{}
	var search = Search{}
	var graphl = Graph{}

	// this is a write transaction

	w := k.Write()

	person := uuid.New()

	object.Set(w, person, "id", person)
	object.Set(w, person, "name", "Bob Baumeister")
	search.Set(w, person, "name", "Bob Baumeister")

	phone := uuid.New()
	object.Set(w, phone, "number", "555-1234")
	graphl.Set(w, person, "phones", phone)

	err = w.Commit(ctx)
	require.NoError(t, err)

	// find bobs phones

	var numbers []string

	r := k.Read()

	for personID, err := range search.Get(r, ctx, "name", "Bob b") {

		require.NoError(t, err)

		var name string
		err := object.Get(r, ctx, personID, "name", &name)
		require.NoError(t, err)
		if name != "Bob Baumeister" {
			continue
		}

		fmt.Println("SEARCH YIELDS ", personID, name)

		for phoneID, err := range graphl.Get(r, ctx, personID, "phones") {
			require.NoError(t, err)

			var number string
			err = object.Get(r, ctx, phoneID, "number", &number)
			require.NoError(t, err)

			numbers = append(numbers, number)

		}
	}

	fmt.Println(numbers)

}
