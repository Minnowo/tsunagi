
PROTOC=protoc

protobuf:
	$(PROTOC) \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		./src/rpc/tsunagi.proto

format:
	gofmt -w -s .
	goimports -w .

