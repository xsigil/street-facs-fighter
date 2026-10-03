BIN_DIR := bin
APP_NAME := $(shell basename $(CURDIR))
IMPORTER_NAME := sff-importer
DB_FILE := app.db

.PHONY: all build run test tidy clean migrate import

all: build

build:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(IMPORTER_NAME) ./cmd/$(IMPORTER_NAME)

run:
	@go run ./cmd/$(APP_NAME)

# migrations/001_init.sql を SQLite に適用
migrate:
	@sqlite3 $(DB_FILE) < migrations/001_init.sql
	@echo "[+] Migration applied to $(DB_FILE)"

# 例: make import SRC=/path/to/cd_root
import:
	@go run ./cmd/$(IMPORTER_NAME) -src "$(SRC)"

test:
	@go test -v ./...

tidy:
	@go mod tidy

clean:
	@rm -rf $(BIN_DIR) *.db *.db-journal assets/*.gif
