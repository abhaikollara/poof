BINARY := poof
MODULE := abhai.dev/poof

.PHONY: build test install clean

build:
	go build -o $(BINARY) ./cmd/poof

test:
	go test ./...

install:
	go install ./cmd/poof
