.PHONY: lint
lint:
	golangci-lint run --build-tags racy -v

.PHONY: lint-fix
lint-fix:
	golangci-lint run --build-tags racy -v --fix

.PHONY: test
test:
	go test -v -count=1 -race -tags unit,integration,racy,long -shuffle on -coverprofile=cover.out ./...

.PHONY: test-unit
test-unit:
	go test -v -count=1 -race -tags unit -shuffle on ./...

.PHONY: test-integration
test-integration:
	go test -v -count=1 -race -tags integration -shuffle on ./...

.PHONY: test-racy
test-racy:
	go test -v -count=1 -race -tags racy -shuffle on ./...

.PHONY: test-long
test-long:
	go test -v -count=1 -race -tags long -shuffle on ./...

.PHONY: cover
cover: test
	@go tool cover -html=cover.out -o=cover.html

.PHONY: messages
messages:
	protoc --go_out=internal/messaging ./internal/messaging/*.proto
	mkdir -p internal/message/protocol
	mkdir -p internal/message/test
	mv internal/messaging/go.eloylp.dev/goomerang/internal/protocol/*.pb.go internal/messaging/protocol
	mv internal/messaging/go.eloylp.dev/goomerang/internal/test/*.pb.go internal/messaging/test
	rm -rf internal/messaging/go.eloylp.dev
