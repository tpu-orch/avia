.PHONY: build run test lint clean docker-up docker-down

build:
	go build -o bin/avia ./cmd/avia

run:
	CONFIG_PATH=config/dev.yaml go run ./cmd/avia

test:
	go test -v ./...

lint:
	golangci-lint run

docker-up:
	docker compose -f config/docker/docker-compose.yml up --build -d

docker-down:
	docker compose -f config/docker/docker-compose.yml down
