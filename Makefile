BINARY  := rootinfo
GO      := go
LDFLAGS := -ldflags="-s -w"

.PHONY: build test lint clean

build:
	$(GO) build $(LDFLAGS) -o $(BINARY) .

test:
	$(GO) test ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)
