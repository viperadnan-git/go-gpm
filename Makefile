.PHONY: build clean install uninstall

PREFIX ?= /usr/local
GO ?= $(shell which go)

build:
	cd cmd/gpcli && $(GO) build -o gpcli .

clean:
	rm -f cmd/gpcli/gpcli

install: build
	install -d $(PREFIX)/bin 
	install -m 755 cmd/gpcli/gpcli $(PREFIX)/bin/gpcli

uninstall:
	rm -f $(PREFIX)/bin/gpcli

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...
	cd cmd/gpcli && go run golang.org/x/tools/cmd/deadcode@latest ./...