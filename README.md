# Test App (HTTP + TCP + UDP)

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

# UDP/TCP Sender

A simple Go application to send UDP or TCP messages for testing servers and clusters.

## Features

* Send messages over **UDP** or **TCP**
* Configurable message content, count, concurrency, and interval
* Optional newline for TCP line-delimited protocols
* Runs as a standalone Go binary or inside Docker
