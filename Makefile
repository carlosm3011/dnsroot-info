BINARY     := rootinfo
VERSION    := 0.3.0
BUILD_DATE := $(shell date -u +%Y-%m-%d)
GO         := go
LDFLAGS    := -ldflags="-s -w -X rootinfo/cmd.Version=$(VERSION) -X rootinfo/cmd.BuildDate=$(BUILD_DATE)"

.PHONY: build test lint clean

build:
	$(GO) build $(LDFLAGS) -o $(BINARY) .

test:
	$(GO) test ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)
