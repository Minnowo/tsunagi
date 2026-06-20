
## tsunagi

tsunagi is the client-server and relay-server implementation for the Pulsar network.

## Running

### Prerequisites

- Go 1.26+
- Docker (if you want to build the wasm module)

### Build

To build the relay:
```sh
go build -o tsunagi .
```

To build the wasm module, make sure you have Docker, and then use the Makefile:
```sh
make tinygo
```

Alternatively read the Makefile and run the commands on your own.

### Usage

**Run the relay server**:
```sh
go run main.go run --http-port 7470 --grpc-port 7471
```

There is a very bad CLI client to talk to relays over GRPC also, but i'm not going to document it now.

