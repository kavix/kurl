APP := .
INSTALL_BIN_DIR ?= $(HOME)/.local/bin

.PHONY: run test build install

run:
	go run $(APP) $(ARGS)

test:
	go test ./...

build:
	go build -ldflags="-s -w" -o kurl $(APP)

install:
	go build -ldflags="-s -w" -o $(INSTALL_BIN_DIR)/kurl $(APP)