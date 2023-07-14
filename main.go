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
	"strconv"
	"syscall"
)

type SheetData = struct {
	Lot    int
	Block  int
	Status string
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

func convert(data [][]string) map[int]string {
	fixed_data := make([]SheetData, 0)
	for _, thing := range data {
		lot, _ := strconv.Atoi(thing[0])
		block, _ := strconv.Atoi(thing[1])

		sdata := SheetData{
			Lot:    lot,
			Block:  block,
			Status: thing[5],
		}

		if sdata.Status != "SOLD" {
			sdata.Status = ""
		}

		fixed_data = append(fixed_data, sdata)
	}

	sort.Slice(fixed_data, func(i, j int) bool {
		// first the lot number, then the block number
		a := fixed_data[i]
		b := fixed_data[j]

		if a.Lot == b.Lot {
			return a.Block < b.Block
		}

		return a.Lot < b.Lot
	})

	result := make(map[int]string, len(fixed_data))
	for i := 0; i < len(fixed_data); i += 1 {
		result[i] = fixed_data[i].Status
	}

	return result
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

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
		data := make([][]string, 115)
		err := json.Unmarshal(buf, &data)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}

		fmt.Printf("number of things: %v\n", len(data))

		Data = convert(data)

		fmt.Printf("new connection from %v\n", r.RemoteAddr)
	}
}
