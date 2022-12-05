SHELL=/usr/bin/env bash

test:
	go test ./internal/{agent,server}/... -coverprofile=coverage.out
	@go tool cover -html=coverage.out

.PHONY: test
