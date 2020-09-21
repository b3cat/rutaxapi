.PHONY: build
build:
	go build -v ./cmd/check

.DEFAULT_GOAL := build