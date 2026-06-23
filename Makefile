.PHONY: fmt vet build run test race docker-build docker-up

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...

build:
	go build -o bin/url-shortener ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

race:
	go test -race ./...

docker-build:
	docker build -t url-shortener .

docker-up:
	docker compose up --build
