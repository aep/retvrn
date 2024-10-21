dev: graph/generated.go
	go build
	./retvrn


graph/generated.go: graph/schema.graphqls gqlgen.yml
	go run github.com/99designs/gqlgen generate


