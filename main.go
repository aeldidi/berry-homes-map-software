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
	"sort"
	"syscall"
)

type SheetData = struct {
	Lot        int
	Block      int
	Address    string
	PocketSize int
	Notes      string
	Closing    string
	Color      string
}

var Data map[int]string

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

func convert(data []SheetData) map[int]string {
	sort.Slice(data, func(i, j int) bool {
		// first the lot number, then the block number
		a := data[i]
		b := data[j]

		if a.Lot == b.Lot {
			return a.Block < b.Block
		}

		return a.Lot < b.Lot
	})

	result := make(map[int]string, len(data))
	for i := 0; i < len(data); i += 1 {
		if data[i].Notes != "SOLD" {
			continue
		}

		result[i] = "SOLD"
	}

	return result
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")

		// FIXME: we should save the data into a file and read it out
		//        here that way if the server needs to be restarted,
		//        you don't have to re-edit the spreadsheet for it to
		//        work.
		if Data == nil {
			w.Write([]byte("[]"))
			return
		}

		data, err := json.Marshal(Data)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}

		_, err = w.Write(data)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
	case http.MethodPost:
		if r.Header.Get("X-I-Am-Silly") != "Yes I am" {
			return
		}

		buf, _ := io.ReadAll(r.Body)

		// FIXME: If the body only has 2 lines and the end of the
		//        second line is a newline, this will crash.
		number_lines := 0
		for i := 0; i < len(buf); i += 1 {
			if number_lines == 2 {
				break
			}

			if buf[i] == byte('\n') {
				number_lines += 1
				buf = buf[i+1:]
			}
		}

		data := make([]SheetData, 115)
		err := json.Unmarshal(buf, &data)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}

		Data = convert(data)

		fmt.Printf("new connection from %v\n", r.RemoteAddr)
	}
}
