package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
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

var IrvineCreekPoints = []canvas.Point{
	// Block 3
	{X: 321, Y: 589}, {X: 334, Y: 566}, {X: 373, Y: 533}, {X: 383, Y: 550}, {X: 393, Y: 566}, {X: 402, Y: 581}, {X: 411, Y: 597}, {X: 418, Y: 611}, {X: 429, Y: 627},

	// Block 4
	{X: 503, Y: 660}, {X: 521, Y: 647}, {X: 534, Y: 636}, {X: 549, Y: 628}, {X: 564, Y: 619}, {X: 581, Y: 609}, {X: 594, Y: 599}, {X: 609, Y: 592}, {X: 629, Y: 585}, {X: 641, Y: 572}, {X: 657, Y: 565}, {X: 698, Y: 541}, {X: 706, Y: 557}, {X: 718, Y: 572}, {X: 726, Y: 585}, {X: 743, Y: 589}, {X: 762, Y: 590}, {X: 779, Y: 590}, {X: 795, Y: 581}, {X: 800, Y: 560}, {X: 800, Y: 542}, {X: 798, Y: 522}, {X: 794, Y: 505}, {X: 786, Y: 488},

	// Block 7
	{X: 717, Y: 470}, {X: 707, Y: 455}, {X: 699, Y: 439}, {X: 691, Y: 421}, {X: 731, Y: 400}, {X: 758, Y: 407}, {X: 778, Y: 406}, {X: 793, Y: 397}, {X: 798, Y: 380}, {X: 796, Y: 363}, {X: 789, Y: 346}, {X: 773, Y: 338}, {X: 756, Y: 336}, {X: 738, Y: 340}, {X: 723, Y: 345}, {X: 701, Y: 354}, {X: 675, Y: 369}, {X: 654, Y: 381}, {X: 634, Y: 395}, {X: 627, Y: 417}, {X: 630, Y: 438}, {X: 640, Y: 451}, {X: 651, Y: 468}, {X: 660, Y: 483}, {X: 668, Y: 497}, {X: 597, Y: 542}, {X: 585, Y: 525}, {X: 573, Y: 511}, {X: 570, Y: 492}, {X: 559, Y: 471}, {X: 543, Y: 471}, {X: 522, Y: 475}, {X: 508, Y: 490}, {X: 505, Y: 513}, {X: 509, Y: 529}, {X: 523, Y: 545}, {X: 534, Y: 557}, {X: 543, Y: 570}, {X: 477, Y: 611}, {X: 466, Y: 594}, {X: 459, Y: 579}, {X: 451, Y: 564}, {X: 441, Y: 550}, {X: 431, Y: 534}, {X: 423, Y: 519}, {X: 416, Y: 502}, {X: 399, Y: 476}, {X: 380, Y: 471}, {X: 359, Y: 467}, {X: 336, Y: 469}, {X: 321, Y: 487}, {X: 317, Y: 504}, {X: 303, Y: 528}, {X: 290, Y: 547}, {X: 280, Y: 566},
}

// There are 112 numbers
var ChurchillMeadowsPoints = []canvas.Point{
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
var Churchill_Meadows []byte

//go:embed website/Irvine_Creek.png
var Irvine_Creek []byte

type Cache struct {
	path string
	imgs map[string][]byte
	c    chan struct {
		name  string
		bytes []byte
	}
}

// Makes a new cache with the specified buffer size
func NewCache(basepath string, buffer int) Cache {
	result := Cache{
		path: basepath,
		imgs: make(map[string][]byte, buffer),
		c: make(
			chan struct {
				name  string
				bytes []byte
			},
			buffer,
		),
	}

	dirs, err := os.ReadDir(result.path)
	if err != nil {
		log.Fatalf("couldn't read '%v': %v\n", basepath, err)
	}

	if len(dirs) == 0 {
		log.Printf("cache is empty\n")
	}

	for _, dir := range dirs {
		if dir.Type().IsDir() {
			log.Printf("directory in cache dir: %v", dir.Name())
			continue
		}

		buf, err := os.ReadFile(path.Join(basepath, dir.Name()))
		if err != nil {
			log.Fatalf("couldn't read file '%v': %v\n", dir.Name(), err)
		}

		log.Printf("reading '%v' from cache\n", dir.Name())
		name := strings.TrimSuffix(dir.Name(), ".png")
		result.imgs[name] = buf
	}

	return result
}

func (c *Cache) WriteImage(name string, w io.Writer) (int, error) {
	return w.Write(c.imgs[name])
}

func (c *Cache) ListenForUpdates() {
	for {
		entry := <-c.c
		path := path.Join(c.path, fmt.Sprintf("%v%v", entry.name, ".png"))
		f, err := os.Create(path)
		if err != nil {
			log.Printf("couldn't write '%v' to cache: %v", entry.name, err)
			continue
		}

		_, err = f.Write(entry.bytes)
		if err != nil {
			log.Printf("couldn't write '%v' to cache: %v", entry.name, err)
			continue
		}

		err = f.Close()
		if err != nil {
			log.Printf("couldn't write '%v' to cache: %v", entry.name, err)
			continue
		}
	}
}

func (c *Cache) Set(name string, img []byte) {
	c.c <- struct {
		name  string
		bytes []byte
	}{name, img}

	c.imgs[name] = img
}

var CacheDir Cache

func init() {
	var dir string
	set := false
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if pair[0] != "REAL_ESTATE_MAP_CACHEDIR" {
			continue
		}

		set = true
		dir = pair[1]
		break
	}

	if !set {
		log.Fatalln("REAL_ESTATE_MAP_CACHEDIR not set!")
	}

	_, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(dir, os.ModeDir)
		if err != nil {
			log.Fatalf("couldn't create directory '%v': %v\n", dir, err)
		}
	} else if err != nil {
		log.Fatalf("couldn't stat directory '%v': %v\n", dir, err)
	}

	CacheDir = NewCache(dir, 2)
}

func main() {
	http.HandleFunc("/churchill-meadow", handler(
		"Churchill_Meadow", ChurchillMeadowsPoints, Churchill_Meadows, 5))
	http.HandleFunc("/irvine-creek",
		handler("Irvine_Creek", IrvineCreekPoints, Irvine_Creek, 6))

	go CacheDir.ListenForUpdates()

	l, err := net.Listen("tcp", ":13370")
	if err != nil {
		log.Fatalf("ERROR couldn't listen on port 13370: %v\n", err)
	}
	defer l.Close()

	// Start the server
	go func() {
		log.Printf("INFO listening at /...\n")
		log.Fatalf("ERROR http.Serve returned with: %v\n",
			http.Serve(l, nil))
	}()

	// Handle common process-killing signals so we can gracefully shut down
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigc
	log.Printf("INFO caught signal %s: shutting down.", sig)
}

func convert(data [][]string, status_column int) map[int]string {
	fixed_data := make([]SheetData, 0)
	for _, thing := range data {
		lot, _ := strconv.Atoi(thing[0])
		block, _ := strconv.Atoi(thing[1])

		sdata := SheetData{
			Lot:    lot,
			Block:  block,
			Status: thing[status_column],
		}

		if sdata.Status != "SOLD" &&
			sdata.Status != "PENDING" &&
			sdata.Status != "ON HOLD" {
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

func handler(
	name string,
	points []canvas.Point,
	input_bytes []byte,
	status_column int,
) func(http.ResponseWriter, *http.Request) {
	input, err := canvas.NewPNGImage(bytes.NewReader(input_bytes))
	if err != nil {
		log.Fatalf("failed to parse PNG image '%v': %v\n", name, err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Cache-Control", "no-cache")

			_, err := CacheDir.WriteImage(name, w)
			if err != nil {
				log.Printf("error: %v\n", err)
				return
			}
		case http.MethodPost:
			if r.Header.Get("X-I-Am-Silly") != "Yes I am" {
				return
			}

			buf, _ := io.ReadAll(r.Body)
			data := [][]string{}
			err := json.Unmarshal(buf, &data)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				return
			}

			log.Printf("%v: new connection from %v\n", name,
				r.RemoteAddr)
			new_data := convert(data, status_column)
			go generateImage(name, points, new_data, input)
		}
	}
}

func generateImage(
	name string,
	points []canvas.Point,
	data map[int]string,
	input canvas.Image,
) {
	red_style := canvas.Style{
		Fill: canvas.Paint{
			Color: canvas.Hex("#ff0000"),
		},
	}
	yellow_style := canvas.Style{
		Fill: canvas.Paint{
			Color: canvas.Hex("#ffd900"),
		},
	}
	green_style := canvas.Style{
		Fill: canvas.Paint{
			Color: canvas.Hex("#42f566"),
		},
	}
	c := canvas.New(float64(input.Bounds().Dx()),
		float64(input.Bounds().Dy()))
	c.RenderImage(input.Image, canvas.Identity)
	for k, v := range data {
		v := strings.ToLower(v)
		point := points[k]
		center := canvas.Identity.Translate(point.X,
			float64(input.Bounds().Dy())-point.Y)
		if strings.Contains(v, "sold") || strings.Contains(v, "closed") {
			c.RenderPath(canvas.Circle(7), red_style, center)
			continue
		}

		if strings.Contains(v, "pending") {
			c.RenderPath(canvas.Circle(7), yellow_style, center)
			continue
		}

		if strings.Contains(v, "on hold") {
			c.RenderPath(canvas.Circle(7), green_style, center)
			continue
		}
	}

	result := rasterizer.Draw(c, canvas.DefaultResolution,
		canvas.SRGBColorSpace{})
	bbuf := &bytes.Buffer{}
	err := png.Encode(bbuf, result)
	if err != nil {
		log.Printf("couldn't encode image: %v\n", err)
		// TODO: write error response here
		return
	}

	CacheDir.Set(name, bbuf.Bytes())
	log.Println("new thing should be ready")
}
