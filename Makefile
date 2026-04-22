BINARY := mehdir
MODULE := abhai.dev/mehdir

.PHONY: build test install clean

build:
	@mkdir -p bin
	go build -o bin/$(BINARY) ./cmd/mehdir

test:
	go test ./...

install:
	go install ./cmd/mehdir

clean:
	rm -rf bin
