.PHONY: build
build:
	go generate ./cmd/monadscli
	go build -o bin/monadscli ./cmd/monadscli
