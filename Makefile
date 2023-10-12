gql:
	go run github.com/99designs/gqlgen generate

tag:
	git tag -a $(t) -m $(m)
	git push origin $(t)

test:
	go test ./...
