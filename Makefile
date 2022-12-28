SHELL=/usr/bin/env bash

init:
	go mod tidy

proto:
	@rm -f internal/pb/*.go
	protoc \
		--proto_path=internal/proto \
		--go_out=internal/pb \
		--go_opt=paths=source_relative \
		--go-grpc_out=internal/pb \
		--go-grpc_opt=paths=source_relative \
	internal/proto/*.proto

test:
	go test ./internal/{agent,server,crypto}/... -coverprofile=coverage.out
	@go tool cover -html=coverage.out

.PHONY: init proto test
