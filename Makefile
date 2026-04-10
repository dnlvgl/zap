BINARY := zap
PKG := ./cmd/zap/

.PHONY: build install test lint clean

build:
	go build -o $(BINARY) $(PKG)

install: build
	install -m 755 $(BINARY) /usr/local/bin/$(BINARY)

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)
