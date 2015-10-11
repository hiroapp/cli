VERSION := $(shell git describe --tags --always --dirty)
default: bin/hiro

bin:
	mkdir -p bin

bin/hiro: bin
	go build -o $@ -ldflags "-X main.version $(VERSION)" ./cmd/hiro

test:
	go test ./...

install: bin/hiro
	install ./bin/hiro /usr/local/bin/

.PHONY: bin/hiro test install
