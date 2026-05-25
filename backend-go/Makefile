.PHONY: all build run test lint clean docker help

APP_NAME=server
BUILD_DIR=./build
MAIN=cmd/server/main.go
GO=go

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'

## all: Build and test the application
all: lint test build

## build: Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)
	@echo "Built: $(BUILD_DIR)/$(APP_NAME)"

## run: Run the application in development mode
run:
	$(GO) run $(MAIN)

## test: Run all tests
test:
	$(GO) test -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run linter
lint:
	golangci-lint run ./...

## fmt: Format Go code
fmt:
	gofmt -w -s .

## vet: Run go vet
vet:
	$(GO) vet ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## docker: Build Docker image
docker:
	docker build -t sino-imgbed-backend:latest .

## docker-run: Run Docker container
docker-run:
	docker run -p 8000:8000 --env-file .env sino-imgbed-backend:latest