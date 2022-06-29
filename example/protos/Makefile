.PHONY: messages
messages:
	protoc --go_out=. *.proto
	mv go.eloylp.dev/goomerang/examples/protos/*.pb.go .
	rm -rf ./go.eloylp.dev
