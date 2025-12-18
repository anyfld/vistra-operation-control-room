.PHONY: build run test test-coverage fmt vet lint generate clean tidy

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

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
