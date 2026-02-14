BINARY := zap
PKG := ./cmd/zap/

.PHONY: build install test clean

build:
	go build -o $(BINARY) $(PKG)

install: build
	install -m 755 $(BINARY) /usr/local/bin/$(BINARY)

test:
	go test ./...

clean:
	rm -f $(BINARY)
