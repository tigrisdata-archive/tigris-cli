VERSION=$(shell git describe --tags --always)
GO_SRC=$(shell find . -name "*.go" -not -name "*_test.go")

BUILD_PARAM=-tags=release -ldflags "-X 'main.Version=$(VERSION)'" $(shell printenv BUILD_PARAM)
TEST_PARAM=-cover -race -tags=test $(shell printenv TEST_PARAM)
export GOPRIVATE=github.com/tigrisdata/tigrisdb-client-go

all: ${GO_SRC}
	go build ${BUILD_PARAM} .

lint:
	golangci-lint run

go.sum: go.mod
	go mod download

test: go.sum lint
	go test $(TEST_PARAM) ./...
