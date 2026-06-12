
DIST  =dist
PROTOC=protoc

protobuf:
	$(PROTOC) \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		./src/rpc/tsunagi.proto

tinygo:
	mkdir -p "${DIST}"
	docker run --rm \
		-v "./mod/tcrypto:/src" \
		-w "/src" \
		-e "GOOS=js" \
		-e "GOARCH=wasm" \
		tinygo/tinygo:0.41.1 \
		sh -c "tinygo build -target=wasm -o tcrypto.wasm wasm/tinygo.go && \
		       cp /usr/local/tinygo/targets/wasm_exec.js ./wasm_exec.js"
	mv ./mod/tcrypto/tcrypto.wasm "${DIST}"
	mv ./mod/tcrypto/wasm_exec.js "${DIST}"

mount-tinygo:
	docker run -it --rm \
		-v "./mod/tcrypto:/src" \
		-w "/src" \
		-e "GOOS=js" \
		-e "GOARCH=wasm" \
		tinygo/tinygo:0.41.1 \
		bash

format:
	gofmt -w -s .
	goimports -w .

