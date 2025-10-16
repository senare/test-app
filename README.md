# Test Server (HTTP + TCP + UDP)

A simple Go server for testing HTTP, TCP, and UDP traffic, suitable for Kubernetes or cluster testing.

## Features

* **HTTP** endpoints:

    * `/` — debug/echo endpoint for request inspection (method, URL, headers, body)
    * `/healthz` — returns `OK` (for Kubernetes health checks)
    * `/version` — returns content of `version.txt`
* **TCP listener** — logs every line/message received
* **UDP listener** — logs every packet received
* **JSON log output** — logs timestamp, protocol, source, message length, and payload
* Runs in Docker for easy deployment

## Ports (default)

| Protocol | Port | Notes                                               |
| -------- | ---- | --------------------------------------------------- |
| HTTP     | 8080 | `/` debug echo, `/healthz` and `/version` endpoints |
| TCP      | 9000 | Logs each line/message received                     |
| UDP      | 9001 | Logs each packet received                           |

## Build (Go)

```bash
# Clone the repo
git clone <repo-url>
cd <repo-folder>

# Build the server
go build -o server server.go

# Run directly
./server
```

## Docker

```bash
# Build the Docker image
docker build -t test-server .

# Run container
docker run --rm -p 8080:8080 -p 9000:9000 -p 9001:9001 test-server

# Run container (host networking for verify test tcp/udp)
docker run -p 0.0.0.0:8080:8080 -p 0.0.0.0:9000:9000 -p 0.0.0.0:9001:9001 test-app:0.1.2
```

## Testing the server

```bash
# HTTP debug/echo
curl -X POST -d "debug test" http://localhost:8080/

# HTTP health/version
curl localhost:8080/healthz
curl localhost:8080/version

# TCP
echo "tcp test" | nc 127.0.0.1 9000

# UDP
echo -n "udp test" | nc -u 127.0.0.1 9001
```

## Logs

All messages (HTTP, TCP, UDP) are logged in JSON lines format to stdout:

```json
{"ts":"2025-10-16T21:22:08.345Z","proto":"UDP","src":"10.0.0.4:51123","len":11,"text":"ping test","hex":"70696e672074657374"}
```

