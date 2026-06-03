.PHONY: build test lint run format

build:
	go build -o twirl ./cmd/twirl/

test:
	go test ./...

lint:
	golangci-lint run
	pnpm run lint:md

format:
	gofmt -w ./...
	goimports -w -local github.com/cajohnson0125/Twirl ./...

run:
	go run ./cmd/twirl/
