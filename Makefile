.PHONY: linux macos docker run build test

NAME ?= envoy-wrapper
VERSION ?= v1.13.0
LDFLAGS ?= -X=main.version=$(VERSION) -w -s
BUILD_FLAGS ?= -v
CGO_ENABLED ?= 0


macos:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=${CGO_ENABLED} go build -o build/macos/${NAME} ${BUILD_FLAGS} -ldflags "$(LDFLAGS)" $^

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=${CGO_ENABLED} go build -o build/linux/${NAME} ${BUILD_FLAGS} -ldflags "$(LDFLAGS)" $^

run: linux
	cp ./build/linux/${NAME} tests/
	cd tests && docker-compose up --build

test:
	go test -v -race $(shell go list ./... )

build: linux
	docker build -t duym/envoy-wrapper:$(VERSION) --build-arg VERSION=$(VERSION) .
	docker push duym/envoy-wrapper:$(VERSION)