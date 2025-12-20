.PHONY: build run test test-coverage fmt vet lint generate clean tidy info sample

build:
	go build -o bin/server ./cmd/server

run:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env from .env.example"; \
	fi
	go run ./cmd/server

info:
	go run ./cmd/info

sample:
	go run ./cmd/sample

test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet
	@if [ ! -f custom-gcl ]; then \
		golangci-lint custom; \
	fi
	./custom-gcl run --fix

generate:
	go generate ./...

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

tidy:
	go mod tidy
