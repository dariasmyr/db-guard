.PHONY: build clean

APP_NAME := build/db-guard

build:
	go build -ldflags="-s -w" -o $(APP_NAME) cmd/db-guard.go

clean:
	rm -f $(APP_NAME)
