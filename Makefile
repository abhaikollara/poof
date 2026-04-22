BINARY := poof
MODULE := abhai.dev/poof

.PHONY: build test install clean

build:
	@mkdir -p bin
	go build -o bin/$(BINARY) ./cmd/poof

test:
	go test ./...

install:
	go install ./cmd/poof

clean:
	rm -rf bin
