.PHONY: test build

GIT_PATH=github.com/derricw
BIN_NAME=siggo
REPO_NAME=siggo
BIN_DIR := $(CURDIR)/bin
INSTALL_DIR := $${HOME}/.local/bin
VERSION  := $(shell git describe --abbrev=0 --tags)
GIT_COMMIT= $(shell git rev-parse HEAD)
BUILD_DATE= $(shell date '+%Y-%m-%d-%H:%M:%S')
GOFMT := gofmt
GO     = go

default: fmt test build release

test:
	go test -v -coverprofile "cov.cov" -covermode=count ./...
	go tool cover -func=cov.cov
	go tool cover -html=cov.cov -o coverage.html

$(BIN_DIR):
	@mkdir -p $@

$(INSTALL_DIR):
	@mkdir -p $@

build: $(BIN_DIR)
	go build -v \
		-ldflags "-X ${GIT_PATH}/${REPO_NAME}/version.GitCommit=${GIT_COMMIT} -X ${GIT_PATH}/${REPO_NAME}/version.BuildDate=${BUILD_DATE}" \
		-o $(BIN_DIR)/$(BIN_NAME)

build_darwin: $(BIN_DIR)
	GOOS=darwin go build -v \
		-ldflags "-X ${GIT_PATH}/${REPO_NAME}/version.GitCommit=${GIT_COMMIT} -X ${GIT_PATH}/${REPO_NAME}/version.BuildDate=${BUILD_DATE}" \
		-o $(BIN_DIR)/$(BIN_NAME)

fmt: ; $(info running gofmt...) @ ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./...); do \
		$(GOFMT) -l -w $$d/*.go || ret$$? ; \
	done ; exit $$ret

install: build $(INSTALL_DIR)
	cp ${BIN_DIR}/${BIN_NAME} ${INSTALL_DIR}/${BIN_NAME}

release_linux: build
	mkdir -p dist
	tar czf dist/siggo-${VERSION}-linux-amd64.tar.gz bin/${BIN_NAME}

release_darwin: build_darwin
	mkdir -p dist
	tar czf dist/siggo-${VERSION}-darwin-amd64.tar.gz bin/${BIN_NAME}

