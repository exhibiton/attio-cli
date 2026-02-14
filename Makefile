BINARY ?= attio
COVER_MIN ?= 70.0

.PHONY: build fmt test test-race test-cover cover-report cover-check integration ci tidy lint

build:
	go build -o bin/$(BINARY) ./cmd/attio

fmt:
	gofmt -w ./cmd ./internal

test:
	go test ./...

test-race:
	go test -race ./...

test-cover:
	go test ./... -coverprofile=coverage.out

cover-report:
	./scripts/coverage_report.sh coverage.out

cover-check:
	./scripts/check_coverage.sh $(COVER_MIN)

integration:
	go test -tags=integration ./internal/integration

lint:
	golangci-lint run

tidy:
	go mod tidy

ci: fmt test test-race cover-check
