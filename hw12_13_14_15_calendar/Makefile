BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(BIN)_scheduler -ldflags "$(LDFLAGS)" ./cmd/scheduler
	go build -v -o $(BIN)_sender -ldflags "$(LDFLAGS)" ./cmd/sender
	go build -v -o $(BIN)_tools -ldflags "$(LDFLAGS)" ./cmd/cli-tools


run: build
	$(BIN) -config ./configs/config.toml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/... ./pkg/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.61.0

lint: install-lint-deps
	golangci-lint -v run ./...

migrate: build
	$(BIN) migrate

generate:
	protoc \
		--proto_path=api \
		--go_out=gen/events/pb \
		--go-grpc_out=gen/events/pb \
		--grpc-gateway_out=gen/events/pb \
		api/events/*.proto

.PHONY: build run build-img run-img version test lint
