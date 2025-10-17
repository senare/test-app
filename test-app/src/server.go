// server.go
package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type LogEntry struct {
	Timestamp string `json:"ts"`
	Proto     string `json:"proto"`
	Src       string `json:"src"`
	Len       int    `json:"len"`
	Text      string `json:"text,omitempty"`
	Hex       string `json:"hex"`
	Path      string `json:"path,omitempty"`
	Method    string `json:"method,omitempty"`
}

// ------------- HTTP handlers -------------

func getHealth(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK\n")
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data, err := os.ReadFile("./version.txt")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func logHTTPRequests(next http.Handler, disableHealthLog bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if disableHealthLog && r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		entry := LogEntry{
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			Proto:     "HTTP",
			Src:       r.RemoteAddr,
			Path:      r.URL.Path,
			Method:    r.Method,
			Len:       int(r.ContentLength),
		}
		b, _ := json.Marshal(entry)
		log.Println(string(b))
		next.ServeHTTP(w, r)
	})
}

// ------------- TCP listener -------------

func startTCPServer(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("TCP listen error: %v", err)
	}
	log.Printf("TCP listening on %s", addr)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				log.Printf("TCP accept error: %v", err)
				continue
			}
			go handleTCPConn(conn)
		}
	}()
}

func handleTCPConn(conn net.Conn) {
	defer conn.Close()
	raddr := conn.RemoteAddr().String()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		data := scanner.Bytes()
		logEntry("TCP", raddr, data)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("TCP read error from %s: %v", raddr, err)
	}
}

// ------------- UDP listener -------------

func startUDPServer(addr string) {
	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Fatalf("UDP listen error: %v", err)
	}
	log.Printf("UDP listening on %s", addr)

	go func() {
		buf := make([]byte, 65535)
		for {
			n, remote, err := pc.ReadFrom(buf)
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				log.Printf("UDP read error: %v", err)
				continue
			}
			data := make([]byte, n)
			copy(data, buf[:n])
			logEntry("UDP", remote.String(), data)
		}
	}()
}

// ------------- Logging helper -------------

func logEntry(proto, src string, data []byte) {
	text := string(data)
	if len(text) > 256 {
		text = text[:256] + "..."
	}
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Proto:     proto,
		Src:       src,
		Len:       len(data),
		Text:      strings.TrimSpace(text),
		Hex:       hex.EncodeToString(data),
	}
	b, _ := json.Marshal(entry)
	log.Println(string(b))
}

// ------------- main -------------

func debugEcho(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "Method: %s\n", r.Method)
	fmt.Fprintf(w, "URL: %s\n", r.URL.String())
	fmt.Fprintf(w, "Proto: %s\n", r.Proto)
	fmt.Fprintf(w, "RemoteAddr: %s\n", r.RemoteAddr)
	fmt.Fprintf(w, "Headers:\n")
	for k, v := range r.Header {
		fmt.Fprintf(w, "  %s: %s\n", k, strings.Join(v, ", "))
	}

	fmt.Fprintf(w, "\nBody:\n")
	body, _ := io.ReadAll(r.Body)
	w.Write(body)
}

// ------------- main -------------

func main() {
	log.SetFlags(0)

	disableHealthLog := os.Getenv("DISABLE_HEALTH_LOG") == "true"

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", debugEcho)
	httpMux.HandleFunc("/healthz", getHealth)
	httpMux.HandleFunc("/version", fileHandler)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: logHTTPRequests(httpMux, disableHealthLog),
	}

	startTCPServer(":9000")
	startUDPServer(":9001")

	go func() {
		log.Println("HTTP listening on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("Shutting down servers...")
	httpServer.Close()
	time.Sleep(500 * time.Millisecond)
	log.Println("Exited cleanly.")
}
