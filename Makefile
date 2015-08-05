default: bin/hiro

bin:
	mkdir -p bin

bin/hiro: bin
	go build -o $@ ./cmd/hiro

docs: docs/edit-format.png

docs/edit-format.png: docs/edit-format.dot
	dot -Tpng -o $@ $^

.PHONY: bin/hiro
