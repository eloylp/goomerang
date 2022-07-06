lint:
	golangci-lint run --build-tags racy -v
lint-fix:
	golangci-lint run --build-tags racy -v --fix
test: test-unit test-integration test-racy test-long
test-unit:
	go test -v -count=1 -race -tags unit -shuffle on ./...
test-integration:
	go test -v -count=1 -race -tags integration -shuffle on ./...
test-racy:
	go test -v -count=1 -race -tags racy -shuffle on ./...
test-long:
	go test -v -count=1 -race -tags long -shuffle on ./...
messages:
	protoc --go_out=internal/messaging ./internal/messaging/*.proto
	mkdir -p internal/message/protocol
	mkdir -p internal/message/test
	mv internal/messaging/go.eloylp.dev/goomerang/internal/protocol/*.pb.go internal/messaging/protocol
	mv internal/messaging/go.eloylp.dev/goomerang/internal/test/*.pb.go internal/messaging/test
	rm -rf internal/messaging/go.eloylp.dev
