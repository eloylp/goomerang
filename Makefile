ALL_TESTS_TAGS=unit,integration,racy,long
DEV_TESTS_TAGS=unit,integration

.DEFAULT_GOAL := all

.PHONY: all
all: lint test

.PHONY: lint
lint:
	golangci-lint run --build-tags ${ALL_TESTS_TAGS} -v

.PHONY: lint-fix
lint-fix:
	golangci-lint run --build-tags ${ALL_TESTS_TAGS} -v --fix

.PHONY: test
test:
	go test -v -race -tags ${DEV_TESTS_TAGS} -shuffle on -coverprofile=cover.out ./...

.PHONY: test-unit
test-unit:
	go test -v -race -tags unit -shuffle on ./...

.PHONY: test-integration
test-integration:
	go test -v -race -tags integration -shuffle on ./...

.PHONY: test-racy
test-racy:
	go test -v -race -tags racy -shuffle on ./...

.PHONY: test-long
test-long:
	go test -v -race -tags long -shuffle on ./...

.PHONY: cover
cover: test test-long test-racy
	@go tool cover -html=cover.out -o=cover.html

.PHONY: messages
messages:
	protoc --go_out=internal/messaging ./internal/messaging/*.proto
	mkdir -p internal/message/protocol
	mkdir -p internal/message/test
	mv internal/messaging/go.eloylp.dev/goomerang/internal/protocol/*.pb.go internal/messaging/protocol
	rm -rf internal/messaging/go.eloylp.dev
