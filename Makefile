lint:
	golangci-lint run --build-tags unit,integration,racy -v
lint-fix:
	golangci-lint run --build-tags unit,integration,racy -v --fix
test:
	go test -v ./...

messages:
	protoc --go_out=internal/message ./internal/message/*.proto
	mkdir -p internal/message/protocol
	mkdir -p internal/message/test
	mv internal/message/go.eloylp.dev/goomerang/internal/protocol/*.pb.go internal/message/protocol
	mv internal/message/go.eloylp.dev/goomerang/internal/test/*.pb.go internal/message/test
	rm -rf internal/message/go.eloylp.dev
