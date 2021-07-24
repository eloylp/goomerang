lint:
	golangci-lint run --build-tags unit,integration,racy -v
lint-fix:
	golangci-lint run --build-tags unit,integration,racy -v --fix
test:
	go test -v ./...

messages:
	protoc --go_out=message ./message/protocol.proto ./message/test.proto
