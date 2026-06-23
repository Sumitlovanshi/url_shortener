.PHONY: run test race docker-build docker-up

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
