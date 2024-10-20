package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"reflect"
	"retvrn/kv"
	"retvrn/kv/index/column"

	"github.com/99designs/gqlgen/graphql"
	"github.com/ettle/strcase"
	// "github.com/vektah/gqlparser/v2/ast"
)

type Resolver struct {
	KV kv.KV
}

func (r *Resolver) resolveModel(ctx context.Context, kvr kv.Read, id uuid.UUID, to interface{}) error {

	t := reflect.TypeOf(to)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("resolveModel usage bug: must be pointer to struct")
	}
	t = t.Elem()
	v := reflect.ValueOf(to).Elem()

	octx := graphql.GetOperationContext(ctx)

	for _, field := range graphql.CollectFieldsCtx(ctx, nil) {

		// FIXME: use the json tags instead to match, or the Position to get it from the graphql source definition
		fn := strcase.ToPascal(field.Definition.Name)
		if fn == "Id" {
			fn = "ID"
		}

		tt, ok := t.FieldByName(fn)
		if !ok {
			fmt.Printf("warning: requested field %s not found in %s\n", fn, t.Name())
			continue
		}

		if tt.Type.Kind() == reflect.Ptr {

			// make a new type and assign to field

			vv := reflect.New(tt.Type.Elem())
			v.FieldByName(tt.Name).Set(vv)

			// TODO recurse here
			for _, field := range graphql.CollectFields(octx, field.Selections, nil) {
				fmt.Println(field.Name)
			}

		} else { // assume its a scalar

			//TODO: use batchget

			v := v.FieldByName(tt.Name).Addr().Interface()
			k := field.ObjectDefinition.Name + ":" + tt.Name
			err := column.Get(kvr, ctx, id, k, v)
			if err != nil {
				return fmt.Errorf("cannot get %s.%s: %s", k, id, err)
			}
		}

	}

	return nil
}
