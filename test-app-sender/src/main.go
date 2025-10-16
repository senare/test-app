// client.go
// Simple sender for testing TCP and UDP servers.
// Usage examples:
//  build: go build -o sender client.go
//  send one UDP packet: ./sender --proto udp --host 127.0.0.1 --port 9001 --message "hello"
//  send 100 TCP messages concurrently: ./sender --proto tcp --host 127.0.0.1 --port 9000 --message "ping" --count 100 --concurrency 10

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	proto := flag.String("proto", "udp", "protocol: udp or tcp")
	host := flag.String("host", "127.0.0.1", "target host")
	port := flag.Int("port", 9001, "target port")
	msg := flag.String("message", "hello", "message to send")
	count := flag.Int("count", 1, "total messages per worker (0 = infinite)")
	workers := flag.Int("concurrency", 1, "number of concurrent workers")
	intervalMs := flag.Int("interval-ms", 0, "interval between messages per worker in milliseconds")
	newline := flag.Bool("newline", false, "append newline to message (useful for TCP line protocols)")
	flag.Parse()

	if *proto != "udp" && *proto != "tcp" {
		log.Fatalf("unsupported proto: %s", *proto)
	}

	payload := []byte(*msg)
	if *newline {
		payload = append(payload, '\n')
	}

	address := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("sending to %s (%s) payload=%q workers=%d count=%d interval=%dms\n", address, *proto, string(payload), *workers, *count, *intervalMs)

	// graceful shutdown
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	stop := make(chan struct{})

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msgsSent := 0
			for {
				select {
				case <-stop:
					return
				default:
				}

				if *proto == "udp" {
					err := sendUDP(address, payload)
					if err != nil {
						log.Printf("worker %d: udp send error: %v", id, err)
					}
				} else {
					err := sendTCP(address, payload)
					if err != nil {
						log.Printf("worker %d: tcp send error: %v", id, err)
					}
				}

				msgsSent++
				if *count > 0 && msgsSent >= *count {
					return
				}
				if *intervalMs > 0 {
					select {
					case <-time.After(time.Duration(*intervalMs) * time.Millisecond):
					case <-stop:
						return
					}
				}
			}
		}(i)
	}

	// wait for signal or workers finish
	finished := make(chan struct{})
	go func() {
		wg.Wait()
		close(finished)
	}()

	select {
	case <-sigc:
		log.Printf("interrupt received, stopping...")
		close(stop)
		wg.Wait()
	case <-finished:
		// done naturally
	}

	log.Printf("sender exiting")
}

func sendUDP(address string, payload []byte) error {
	conn, err := net.Dial("udp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(payload)
	return err
}

func sendTCP(address string, payload []byte) error {
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(payload)
	return err
}
