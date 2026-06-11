GO := go
GOFLAGS := -v -race

.PHONY: test test-short lint vet fmt build tidy examples cover-html

test:
	$(GO) test $(GOFLAGS) -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

test-short:
	$(GO) test -short ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -s -w .

build:
	$(GO) build ./...

tidy:
	$(GO) mod tidy

examples:
	$(GO) build ./examples/...

cover-html:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

.DEFAULT_GOAL := test
