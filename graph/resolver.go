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
	"retvrn/kv/index/graph"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/ettle/strcase"
	"github.com/vektah/gqlparser/v2/ast"
)

type Resolver struct {
	KV kv.KV
}

type recursion struct {
	fields ast.SelectionSet
	debug  string
}

func (r *Resolver) resolveModel(ctx context.Context, kvr kv.Read, id uuid.UUID, to interface{}, rec *recursion) error {

	debug := ""
	if rec != nil {
		debug = rec.debug
	}

	t := reflect.TypeOf(to)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("resolveModel usage bug: must be pointer to struct")
	}
	t = t.Elem()
	v := reflect.ValueOf(to).Elem()

	octx := graphql.GetOperationContext(ctx)

	var fields []graphql.CollectedField
	if rec == nil {
		fields = graphql.CollectFieldsCtx(ctx, nil)
	} else {
		fields = graphql.CollectFields(octx, rec.fields, nil)
	}

	for _, field := range fields {

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

		if tt.Type.Kind() == reflect.Slice {
			ttt := tt.Type.Elem()
			if ttt.Kind() != reflect.Ptr {
				return fmt.Errorf("[%s] slice of scalar not supported yet")
			}
			k := field.ObjectDefinition.Name + ":" + tt.Name

			for id2, err := range graph.GetN(kvr, ctx, id, k) {
				if err != nil {
					return fmt.Errorf("[%s] cannot follow graph %s from %s id %s: %s", debug, field.ObjectDefinition.Name, tt.Name, id, err)
				}
				vv := reflect.New(ttt.Elem())
				err = r.resolveModel(ctx, kvr, id2, vv.Interface(), &recursion{
					fields: field.Selections,
				})
				if err != nil {
					return fmt.Errorf("[%s] %w", debug, err)
				}
				v.FieldByName(tt.Name).Set(reflect.Append(v.FieldByName(tt.Name), vv))
			}

		} else if tt.Type.Kind() == reflect.Ptr {

			// make a new type and assign to field

			vv := reflect.New(tt.Type.Elem())
			v.FieldByName(tt.Name).Set(vv)

			k := field.ObjectDefinition.Name + ":" + tt.Name

			id2, err := graph.Get1(kvr, ctx, id, k)
			if err != nil {
				return fmt.Errorf("[%s] cannot follow graph %s from %s id %s: %s", debug, field.ObjectDefinition.Name, tt.Name, id, err)
			}

			err = r.resolveModel(ctx, kvr, id2, vv.Interface(), &recursion{
				fields: field.Selections,
			})
			if err != nil {
				return fmt.Errorf("[%s] %w", debug, err)
			}

		} else { // assume its a scalar

			//TODO: use batchget

			v := v.FieldByName(tt.Name).Addr().Interface()
			k := field.ObjectDefinition.Name + ":" + tt.Name
			_, err := column.Get(kvr, ctx, id, k, v)
			if err != nil {
				if !strings.Contains(err.Error(), "not exist") {
					return fmt.Errorf("[%s] cannot get %s.%s: %s", debug, k, id, err)
				}
			}
		}

	}

	return nil
}
