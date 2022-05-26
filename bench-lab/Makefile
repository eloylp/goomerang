
.PHONY: bench-lab
bench-lab:
	CGO_ENABLED=0 go build -trimpath -tags timetzdata -o ./server/server server/main.go
	CGO_ENABLED=0 go build -trimpath -tags timetzdata -o ./proxy/proxy proxy/main.go
	CGO_ENABLED=0 go build -trimpath -tags timetzdata -o ./client/client client/main.go
	docker-compose up

.PHONY: messages
messages:
	protoc --go_out=model/ ./model/*.proto
	mv model/go.eloylp.dev/goomerang/bench-lab/model/*.pb.go ./model
	rm -rf model/go.eloylp.dev
