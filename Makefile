include .env


TEST_FLAGS ?= -v -race -parallel 5 -shuffle=on
COVER_FLAGS ?= -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
BINARY_NAME ?= bin/app

.PHONY: docker clean-docker test lint deps build clean mock
.DEFAULT_GOAL := help

docker: clean-docker
	docker compose up -d --build

clean-docker:
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
	go build -o $(BINARY_NAME) ./cmd/

clean:
	rm -rf bin/
	rm -f cover.out
mock:
	mockery
help:
	@echo "Available targets:"
	@echo "  docker      - Rebuild and restart docker containers"
	@echo "  clean-docker - Clean docker containers and images"
	@echo "  test        - Run tests with race detection and coverage"
	@echo "  lint        - Run golangci-lint"
	@echo "  deps        - Download dependencies"
	@echo "  build       - Build application"
	@echo "  clean       - Clean build artifacts"
	@echo "  mock        - Generate mocks using mockery with config"