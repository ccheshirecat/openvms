SHELL := /bin/bash

.PHONY: fmt test

fmt:
	gofmt -s -w spec/cli/ovm/main.go || true

test:
	cd tests/conformance && go test -v ./...
