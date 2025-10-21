# Makefile
BINARY_NAME=gophkeeper
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD)
DATE?=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

.PHONY: build
build:
	go build -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o bin/$(BINARY_NAME) ./cmd/client

.PHONY: build-all
build-all: build-linux build-windows build-darwin

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/client

.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/client

.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/client

.PHONY: install
install:
	go install -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" ./cmd/client

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: docker-build
docker-build:
	docker build -t gophkeeper-client:$(VERSION) .

.PHONY: release
release: build-all
	tar -czf bin/$(BINARY_NAME)-linux-amd64.tar.gz -C bin $(BINARY_NAME)-linux-amd64
	zip -j bin/$(BINARY_NAME)-windows-amd64.zip bin/$(BINARY_NAME)-windows-amd64.exe
	tar -czf bin/$(BINARY_NAME)-darwin-amd64.tar.gz -C bin $(BINARY_NAME)-darwin-amd64