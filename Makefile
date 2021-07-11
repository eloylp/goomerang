

lint:
	golangci-lint run --build-tags unit,integration,racy -v
lint-fix:
	golangci-lint run --build-tags unit,integration,racy -v --fix
test:
	go test -v ./...