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
	{X: 657, Y: 1194}, {X: 679, Y: 1140}, {X: 745, Y: 1081},
	{X: 772, Y: 1116}, {X: 792, Y: 1145}, {X: 812, Y: 1178},
	{X: 827, Y: 1208}, {X: 844, Y: 1241}, {X: 860, Y: 1271},

	// Block 4
	{X: 1016, Y: 1334}, {X: 1053, Y: 1311}, {X: 1084, Y: 1291},
	{X: 1119, Y: 1268}, {X: 1141, Y: 1254}, {X: 1177, Y: 1231},
	{X: 1201, Y: 1214}, {X: 1241, Y: 1197}, {X: 1266, Y: 1181},
	{X: 1302, Y: 1162}, {X: 1331, Y: 1146}, {X: 1407, Y: 1097},
	{X: 1428, Y: 1132}, {X: 1450, Y: 1163}, {X: 1471, Y: 1187},
	{X: 1505, Y: 1194}, {X: 1544, Y: 1193}, {X: 1582, Y: 1186},
	{X: 1604, Y: 1163}, {X: 1613, Y: 1132}, {X: 1611, Y: 1097},
	{X: 1610, Y: 1059}, {X: 1601, Y: 1025}, {X: 1588, Y: 984},

	// Block 7
	{X: 1459, Y: 950}, {X: 1434, Y: 913}, {X: 1421, Y: 888},
	{X: 1399, Y: 849}, {X: 1482, Y: 809}, {X: 1525, Y: 817},
	{X: 1571, Y: 815}, {X: 1600, Y: 800}, {X: 1614, Y: 772},
	{X: 1618, Y: 734}, {X: 1602, Y: 711}, {X: 1571, Y: 686},
	{X: 1535, Y: 687}, {X: 1496, Y: 691}, {X: 1462, Y: 701},
	{X: 1424, Y: 720}, {X: 1366, Y: 756}, {X: 1326, Y: 779},
	{X: 1289, Y: 805}, {X: 1281, Y: 846}, {X: 1281, Y: 885},
	{X: 1302, Y: 912}, {X: 1324, Y: 947}, {X: 1344, Y: 976},
	{X: 1358, Y: 1009}, {X: 1204, Y: 1095}, {X: 1184, Y: 1067},
	{X: 1169, Y: 1031}, {X: 1155, Y: 991}, {X: 1134, Y: 963},
	{X: 1099, Y: 954}, {X: 1057, Y: 963}, {X: 1039, Y: 1000},
	{X: 1035, Y: 1039}, {X: 1043, Y: 1069}, {X: 1068, Y: 1093},
	{X: 1095, Y: 1122}, {X: 1110, Y: 1153}, {X: 972, Y: 1234},
	{X: 952, Y: 1203}, {X: 933, Y: 1174}, {X: 916, Y: 1140},
	{X: 898, Y: 1112}, {X: 877, Y: 1077}, {X: 860, Y: 1049},
	{X: 841, Y: 1019}, {X: 812, Y: 965}, {X: 765, Y: 958},
	{X: 724, Y: 950}, {X: 682, Y: 956}, {X: 649, Y: 986},
	{X: 635, Y: 1023}, {X: 616, Y: 1074}, {X: 596, Y: 1110},
	{X: 558, Y: 1150},
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

		if sdata.Status != "SOLD" && sdata.Status != "PENDING" {
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

func handler(name string, points []canvas.Point, input_bytes []byte, status_column int) func(http.ResponseWriter, *http.Request) {
	input, err := canvas.NewPNGImage(bytes.NewReader(input_bytes))
	if err != nil {
		log.Fatalf("failed to parse PNG image '%v': %v\n", name, err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v: request: %#v", name, r)
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

			log.Printf("%v: new connection from %v\n", name, r.RemoteAddr)
			new_data := convert(data, status_column)
			go generateImage(name, points, new_data, input)
		}
	}
}

func generateImage(name string, points []canvas.Point, data map[int]string, input canvas.Image) {
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
	c := canvas.New(float64(input.Bounds().Dx()),
		float64(input.Bounds().Dy()))
	c.RenderImage(input.Image, canvas.Identity)
	for k, v := range data {
		point := points[k]
		center := canvas.Identity.Translate(point.X,
			float64(input.Bounds().Dy())-point.Y)
		switch v {
		case "SOLD":
			c.RenderPath(canvas.Circle(7), red_style, center)
		case "PENDING":
			c.RenderPath(canvas.Circle(7), yellow_style, center)
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
