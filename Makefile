.PHONY: all build run test test-cover benchmark lint docker-build docker-up docker-down clean

BINARY_NAME=bin/mediapulse

all: test build

build:
	@mkdir -p bin
	go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/server

run: build
	./$(BINARY_NAME)

test:
	go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

benchmark:
	go test -bench=. -benchmem ./internal/...

lint:
	go vet ./...

docker-build:
	docker build -t mediapulse-ai:latest .

docker-up:
	docker compose up -d

docker-down:
	docker compose down -v

clean:
	rm -rf bin/ coverage.out coverage.html
