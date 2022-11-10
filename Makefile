VERSION=$(shell git describe --tags --always)
GO_SRC=$(shell find . -name "*.go" -not -name "*_test.go")
BIN=tigris
ifeq ($(GOOS), windows)
BIN=tigris.exe
endif

BUILD_PARAM=-tags=release -ldflags "-w -extldflags '-static' -X 'github.com/tigrisdata/tigris-cli/util.Version=$(VERSION)'" -o ${BIN} $(shell printenv BUILD_PARAM)
TEST_PARAM=-cover -race -tags=test $(shell printenv TEST_PARAM)

all: ${BIN}

${BIN}: ${GO_SRC} go.sum
	CGO_ENABLED=0 go build ${BUILD_PARAM} .

lint:
	golangci-lint run --fix
	shellcheck tests/*.sh
	cd pkg/npm && npm i; npx eslint install.js

go.sum: go.mod
	go mod download

test: ${BIN} go.sum
	go test $(TEST_PARAM) ./...
	/bin/bash tests/*.sh

install: ${BIN}
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 755 ${BIN} $(DESTDIR)$(PREFIX)/bin/

clean:
	rm -f ${BIN}
