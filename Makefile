lint:
	golangci-lint run --build-tags unit,integration,racy -v
lint-fix:
	golangci-lint run --build-tags unit,integration,racy -v --fix
test:
	go test -v ./...

messages:
	protoc --go_out=internal/messaging ./internal/messaging/*.proto
	mkdir -p internal/message/protocol
	mkdir -p internal/message/test
	mv internal/messaging/go.eloylp.dev/goomerang/internal/protocol/*.pb.go internal/messaging/protocol
	mv internal/messaging/go.eloylp.dev/goomerang/internal/test/*.pb.go internal/messaging/test
	rm -rf internal/messaging/go.eloylp.dev
