package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"

	_ "embed"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
)

type SheetData struct {
	Lot    int
	Block  int
	Status string
}

// There are 112 numbers
var Points = []canvas.Point{
	// First block
	{X: 321, Y: 153}, {X: 384, Y: 150}, {X: 445, Y: 149},
	{X: 500, Y: 152}, {X: 548, Y: 156}, {X: 591, Y: 159},
	{X: 604, Y: 229}, {X: 617, Y: 186}, {X: 634, Y: 148},
	{X: 723, Y: 144}, {X: 793, Y: 157}, {X: 804, Y: 208},
	{X: 799, Y: 255}, {X: 873, Y: 266}, {X: 874, Y: 246},
	{X: 864, Y: 221}, {X: 858, Y: 202}, {X: 827, Y: 161},
	{X: 857, Y: 143}, {X: 890, Y: 139}, {X: 923, Y: 141},
	{X: 947, Y: 163}, {X: 960, Y: 193}, {X: 930, Y: 219},
	{X: 926, Y: 241}, {X: 925, Y: 261}, {X: 929, Y: 283},
	{X: 927, Y: 314}, {X: 926, Y: 335}, {X: 927, Y: 355},
	{X: 928, Y: 373}, {X: 927, Y: 396}, {X: 930, Y: 417},

	// Second Block
	{X: 284, Y: 273}, {X: 309, Y: 273}, {X: 328, Y: 273},
	{X: 351, Y: 271}, {X: 373, Y: 271}, {X: 393, Y: 272},
	{X: 412, Y: 273}, {X: 432, Y: 275}, {X: 453, Y: 279},
	{X: 472, Y: 284}, {X: 492, Y: 289}, {X: 511, Y: 295},
	{X: 530, Y: 303}, {X: 559, Y: 314}, {X: 577, Y: 326},
	{X: 598, Y: 334}, {X: 620, Y: 341}, {X: 642, Y: 345},
	{X: 666, Y: 348}, {X: 690, Y: 350}, {X: 710, Y: 351},
	{X: 731, Y: 351}, {X: 751, Y: 351}, {X: 774, Y: 352},
	{X: 794, Y: 354}, {X: 811, Y: 352}, {X: 834, Y: 355},
	{X: 855, Y: 354}, {X: 876, Y: 356}, {X: 877, Y: 378},
	{X: 852, Y: 377}, {X: 830, Y: 379}, {X: 808, Y: 375},
	{X: 782, Y: 373}, {X: 760, Y: 373}, {X: 738, Y: 372},
	{X: 714, Y: 371}, {X: 693, Y: 373}, {X: 670, Y: 371},
	{X: 650, Y: 370}, {X: 630, Y: 368}, {X: 608, Y: 361},
	{X: 588, Y: 353}, {X: 568, Y: 345}, {X: 548, Y: 337},
	{X: 517, Y: 322}, {X: 496, Y: 312}, {X: 471, Y: 308},
	{X: 447, Y: 302}, {X: 423, Y: 298}, {X: 399, Y: 295},
	{X: 373, Y: 295}, {X: 351, Y: 294}, {X: 330, Y: 294},
	{X: 305, Y: 296}, {X: 284, Y: 294},

	// Third Block
	{X: 310, Y: 421}, {X: 333, Y: 420}, {X: 357, Y: 422},
	{X: 382, Y: 420}, {X: 407, Y: 422}, {X: 427, Y: 424},
	{X: 449, Y: 430}, {X: 470, Y: 436}, {X: 502, Y: 452},
	{X: 523, Y: 461}, {X: 543, Y: 470}, {X: 570, Y: 477},
	{X: 595, Y: 483}, {X: 620, Y: 490}, {X: 646, Y: 491},
	{X: 670, Y: 494}, {X: 692, Y: 495}, {X: 714, Y: 496},
	{X: 736, Y: 498}, {X: 760, Y: 500}, {X: 784, Y: 498},
	{X: 806, Y: 501}, {X: 831, Y: 501}, {X: 849, Y: 500},
}

//go:embed website/Churchill_Meadows.png
var _Image []byte
var PreviousPath string
var Previous canvas.Image
var Image canvas.Image

func init() {
	var err error

	set := false
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if pair[0] != "REAL_ESTATE_MAP_CACHEFILE" {
			continue
		}

		set = true
		PreviousPath = pair[1]
		break
	}

	if !set {
		panic("REAL_ESTATE_MAP_CACHEFILE not set!")
	}

	Image, err = canvas.NewPNGImage(bytes.NewReader(_Image))
	if err != nil {
		panic("couldn't read Churchill_Meadows.png")
	}

	f, err := os.Open(PreviousPath)
	if err != nil {
		panic(fmt.Sprintf("couldn't open %v: %v", PreviousPath, err))
	}
	defer f.Close()

	Previous, err = canvas.NewPNGImage(f)
	if err != nil {
		panic(fmt.Sprintf("couldn't read %v: %v", PreviousPath, err))
	}
}

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
		log.Fatalf("ERROR http.Serve returned with: %v",
			http.Serve(l, nil))
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

		if a.Block == b.Block {
			return a.Lot < b.Lot
		}

		return a.Block < b.Block
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
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-cache")

		_, err := w.Write(Previous.Bytes)
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

		log.Printf("new connection from %v\n", r.RemoteAddr)
		new_data := convert(data)
		go generateImage(new_data)
		w.WriteHeader(200)
	}
}

func generateImage(data map[int]string) {
	style := canvas.Style{
		Fill: canvas.Paint{
			Color: canvas.Hex("#ff0000"),
		},
	}
	c := canvas.New(float64(Image.Bounds().Dx()),
		float64(Image.Bounds().Dy()))
	c.RenderImage(Image.Image, canvas.Identity)
	for k, v := range data {
		if v != "SOLD" {
			continue
		}

		point := Points[k]
		center := canvas.Identity.Translate(point.X,
			float64(Image.Bounds().Dy())-point.Y)

		// draw the circle at the point
		c.RenderPath(canvas.Circle(5), style, center)
	}

	result := rasterizer.Draw(c, canvas.DefaultResolution,
		canvas.SRGBColorSpace{})
	// Save the previous image
	f, err := os.Create(PreviousPath)
	if err != nil {
		log.Printf("couldn't open file: %v\n", err)
		return
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Printf("couldn't write file: %v\n", err)
		}
	}()

	bbuf := &bytes.Buffer{}
	err = png.Encode(io.MultiWriter(f, bbuf), result)
	if err != nil {
		log.Printf("couldn't encode image: %v\n", err)
		// TODO: write error response here
		return
	}

	Previous, err = canvas.NewPNGImage(bytes.NewReader(bbuf.Bytes()))
	if err != nil {
		log.Printf("couldn't encode image: %v\n", err)
		// TODO: write error response here
		return
	}
	log.Println("new thing should be ready")
}
