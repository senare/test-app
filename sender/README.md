# UDP/TCP Sender

A simple Go application to send UDP or TCP messages for testing servers and clusters.

## Features

* Send messages over **UDP** or **TCP**
* Configurable message content, count, concurrency, and interval
* Optional newline for TCP line-delimited protocols
* Runs as a standalone Go binary or inside Docker

## Build (Go)

```bash
# Clone the repo
git clone <repo-url>
cd <repo-folder>/src

# Build the sender
go build -o sender client.go

# Run directly
./sender --proto udp --host 127.0.0.1 --port 9001 --message "hello"
```

## Docker

```bash
# Build the Docker image
docker build -t sender:0.0.1 .

# Run UDP sender
docker run --rm --network host sender:0.0.1 --proto udp --host 127.0.0.1 --port 9001 --message "hello from docker"

# Run TCP sender
docker run --rm  --network host sender:0.0.1 --proto tcp --host 127.0.0.1 --port 9000 --message "tcp test" --newline
```

### Optional flags

| Flag            | Description                                   | Default     |
| --------------- | --------------------------------------------- | ----------- |
| `--proto`       | Protocol: `udp` or `tcp`                      | `udp`       |
| `--host`        | Target host/IP                                | `127.0.0.1` |
| `--port`        | Target port                                   | `9001`      |
| `--message`     | Message payload                               | `hello`     |
| `--count`       | Number of messages per worker (0=infinite)    | `1`         |
| `--concurrency` | Number of concurrent workers                  | `1`         |
| `--interval-ms` | Delay between messages per worker (ms)        | `0`         |
| `--newline`     | Append newline to message (TCP line protocol) | `false`     |

## Example usage

```bash
# Send 10 UDP messages to 127.0.0.1:9001
docker run --rm sender --proto udp --host 127.0.0.1 --port 9001 --message "ping" --count 10

# Send 50 TCP messages concurrently using 5 workers
docker run --rm sender --proto tcp --host 127.0.0.1 --port 9000 --message "hello" --count 10 --concurrency 5 --newline
```
