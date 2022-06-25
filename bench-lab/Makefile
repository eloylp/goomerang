
.PHONY: up
up:
	CGO_ENABLED=0 go build -trimpath -tags timetzdata -o ./server/server server/main.go
	CGO_ENABLED=0 go build -trimpath -tags timetzdata -o ./proxy/proxy proxy/main.go
	CGO_ENABLED=0 go build -trimpath -tags timetzdata -o ./client/client client/main.go
	docker-compose up -d --force-recreate
	@echo "Grafana access at: http://localhost:3000"
	@echo "Prometheus access at: http://localhost:3000"
	@echo "Server pprof access at: http://localhost:6060/debug/pprof"
	@echo "Client pprof access at: http://localhost:6061/debug/pprof"

.PHONY: down
down:
	docker-compose down -v

.PHONY: logs
logs:
	docker-compose logs -f

.PHONY: update
update:
	cd server && go get -u go.eloylp.dev/goomerang && cd ..
	cd client && go get -u go.eloylp.dev/goomerang && cd ..

.PHONY: messages
messages:
	protoc --go_out=model/ ./model/*.proto
	mv model/go.eloylp.dev/goomerang/bench-lab/model/*.pb.go ./model
	rm -rf model/go.eloylp.dev
