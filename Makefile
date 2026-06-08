BINARY  := rootinfo
VERSION := 0.2
GO      := go
LDFLAGS := -ldflags="-s -w -X rootinfo/cmd.Version=$(VERSION)"

.PHONY: build test lint clean

build:
	$(GO) build $(LDFLAGS) -o $(BINARY) .

test:
	$(GO) test ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)
