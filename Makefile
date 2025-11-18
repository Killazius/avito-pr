include .env


TEST_FLAGS ?= -v -race -parallel 5 -shuffle=on
COVER_FLAGS ?= -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
BINARY_NAME ?= bin/app

.PHONY: docker-up docker-clean test lint deps build clean mock docker-build
.DEFAULT_GOAL := help

docker:
	docker compose up -d

docker-build:
	docker compose build

docker-clean:
	docker compose down
	docker image prune -f

test:
	go test $(TEST_FLAGS) $(COVER_FLAGS) ./...

lint:
	golangci-lint run ./...

deps:
	go mod download
	go mod verify
	go mod tidy

build: deps
	go build -o $(BINARY_NAME) ./cmd/app

clean:
	rm -rf bin/
	rm -f cover.out
mock:
	mockery
help:
	@echo "Available targets:"
	@echo "  docker      - Start docker containers"
	@echo "  docker-clean - Clean docker containers and images"
	@echo "  docker-build - Build docker images"
	@echo "  test        - Run tests with race detection and coverage"
	@echo "  lint        - Run golangci-lint"
	@echo "  deps        - Download dependencies"
	@echo "  build       - Build application"
	@echo "  clean       - Clean build artifacts"
	@echo "  mock        - Generate mocks using mockery with config"