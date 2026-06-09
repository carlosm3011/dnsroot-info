BINARY     := rootinfo
VERSION    := 0.4.0
BUILD_DATE := $(shell date -u +%Y-%m-%d)
GO         := go
LDFLAGS    := -ldflags="-s -w -X rootinfo/cmd.Version=$(VERSION) -X rootinfo/cmd.BuildDate=$(BUILD_DATE)"
DIST_DIR   := dist

.PHONY: build test lint clean dist

build:
	$(GO) build $(LDFLAGS) -o $(BINARY) .

dist:
	mkdir -p $(DIST_DIR)
	GOOS=darwin  GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/rootinfo-darwin-arm64      .
	GOOS=linux   GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/rootinfo-linux-amd64       .
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/rootinfo-windows-amd64.exe .

test:
	$(GO) test ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)
	rm -rf $(DIST_DIR)
