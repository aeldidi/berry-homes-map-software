package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var Data [][]string

func main() {
	http.HandleFunc("/", handler)

	l, err := net.Listen("tcp", ":13370")
	if err != nil {
		log.Fatalf("ERROR couldn't listen on port 13370: %v", err)
	}
	defer l.Close()

	// Start the server
	go func() {
		log.Printf("INFO listening at /...")
		log.Fatalf("ERROR http.Serve returned with: %v", http.Serve(l, nil))
	}()

	// Handle common process-killing signals so we can gracefully shut down
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigc
	log.Printf("INFO caught signal %s: shutting down.", sig)
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, `
		<!DOCTYPE html>
		<head>
		</head>
		<body>
			<code>%v</code>
		</body>
		`, Data)
	case http.MethodPost:
		buf, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(buf, &Data)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}

		fmt.Printf("new connection from %v: %v\n", r.RemoteAddr, Data[0][0])
	}
}
