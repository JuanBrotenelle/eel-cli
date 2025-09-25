# Makefile for eel-cli

.PHONY: build clean test install

build:
	go build -o eel.exe cmd/eel-cli/main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o eel-cli-linux cmd/eel-cli/main.go

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o eel-cli-darwin cmd/eel-cli/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o eel.exe cmd/eel-cli/main.go

build-all: build-linux build-darwin build-windows

clean:
	rm -f eel-cli.exe eel-cli-linux eel-cli-darwin eel-cli-windows.exe

test:
	go test ./...

install:
	go mod tidy
	go mod download

run:
	go run cmd/eel-cli/main.go

fmt:
	go fmt ./...

help:
	@echo "Available targets:"
	@echo "  build        - Build the CLI for current platform"
	@echo "  build-linux  - Build for Linux"
	@echo "  build-darwin - Build for macOS"
	@echo "  build-windows- Build for Windows"
	@echo "  build-all    - Build for all platforms"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run tests"
	@echo "  install      - Install dependencies"
	@echo "  run          - Run the CLI"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  help         - Show this help"
