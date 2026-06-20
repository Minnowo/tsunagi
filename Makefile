
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
		sh -c "tinygo build -target=wasm -o tcrypto.wasm ./wasm && \
		       cp /usr/local/tinygo/targets/wasm_exec.js ./wasm_exec.js"
	mv ./mod/tcrypto/tcrypto.wasm "${DIST}"
	mv ./mod/tcrypto/wasm_exec.js "${DIST}"
	# see https://github.com/tinygo-org/tinygo/issues/5357
	# The runtime.getRandomData function was removed, and we need to add it back.
	sed -i 's|// func sleepTicks(timeout int64)|"runtime.getRandomData": (slice_ptr, slice_len, slice_cap) => {\n\t\t\t\t\t\tconst buf = loadSlice(slice_ptr, slice_len, slice_cap);\n\t\t\t\t\t\tfor (let offset = 0; offset < buf.length; offset += 65536) {\n\t\t\t\t\t\t\tcrypto.getRandomValues(buf.subarray(offset, Math.min(offset + 65536, buf.length)));\n\t\t\t\t\t\t}\n\t\t\t\t\t},\n\t\t\t\t\t// func sleepTicks(timeout int64)|' "${DIST}/wasm_exec.js"

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

