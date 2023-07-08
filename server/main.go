package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	http.HandleFunc("/", handler)

	l, err := net.Listen("tcp", ":1337")
	if err != nil {
		log.Fatalf("ERROR couldn't listen on port 1337: %v", err)
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
		fmt.Fprintln(w, "coolio")
	case http.MethodPost:
		buf, _ := io.ReadAll(r.Body)
		fmt.Printf("%v", buf)
	}
}
