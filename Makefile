default: bin/hiro

bin:
	mkdir -p bin

bin/hiro: bin
	go build -o $@ ./cmd/hiro

test:
	go test ./...

.PHONY: bin/hiro test
