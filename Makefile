VERSION=$(shell git describe --tags --always)
GO_SRC=$(shell find . -name "*.go" -not -name "*_test.go")
BIN=tigris
ifeq ($(GOOS), windows)
BIN=tigris.exe
endif

BUILD_PARAM=-tags=release -ldflags "-w -extldflags '-static' -X 'github.com/tigrisdata/tigrisdb-cli/util.Version=$(VERSION)'" -o ${BIN} $(shell printenv BUILD_PARAM)
TEST_PARAM=-cover -race -tags=test $(shell printenv TEST_PARAM)
export GOPRIVATE=github.com/tigrisdata/tigrisdb-client-go

all: ${BIN}

${BIN}: ${GO_SRC} go.sum
	CGO_ENABLED=0 go build ${BUILD_PARAM} .

lint:
	golangci-lint run

go.sum: go.mod
	go mod download

test: ${BIN} go.sum
	go test $(TEST_PARAM) ./...
	/bin/bash tests/*.sh

clean:
	rm -f ${BIN}
