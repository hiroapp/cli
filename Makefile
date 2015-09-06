default: bin/hiro

bin:
	mkdir -p bin

bin/hiro: bin
	go build -o $@ ./cmd/hiro

test:
	go test ./...

install: bin/hiro
	install ./bin/hiro /usr/local/bin/

.PHONY: bin/hiro test install
