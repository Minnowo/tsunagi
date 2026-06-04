
## tsunagi

tsunagi is the client-server and relay-server implementation for the Pulsar network.

## Running

### Prerequisites

- Go 1.26+

### Build

```sh
go build -o tsunagi .
```

### Usage

**Initialize a session identity:**
```sh
./tsunagi session init
```

**Run the relay server** (default: `0.0.0.0:7471`):
```sh
./tsunagi run
./tsunagi run --host 0.0.0.0 --port 7471
```

**Run the client WebSocket server** (default: `0.0.0.0:8080`):
```sh
./tsunagi client-server
./tsunagi client-server --host 0.0.0.0 --port 8080
```

**Connect a CLI client to the relay server:**
```sh
./tsunagi client connect --addr tcp://localhost:7471/
./tsunagi client connect --addr tcp://localhost:7471/ --device <device-id>
```

**Connect a relay node:**
```sh
./tsunagi relay connect --addr tcp://localhost:7471/
./tsunagi relay connect --addr tcp://localhost:7471/ --device <device-id>
```


