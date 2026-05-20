.PHONY: swagger run test

swag:
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

run:
	go run ./cmd/server

test:
	go test ./...
