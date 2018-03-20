package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	m  = make(map[string]int)
	cs = map[byte][]int{
		'0': []int{100, 300},
		'1': []int{0, 0},
		'2': []int{100, 0},
		'3': []int{200, 0},
		'4': []int{0, 100},
		'5': []int{100, 100},
		'6': []int{200, 100},
		'7': []int{0, 200},
		'8': []int{100, 200},
		'9': []int{200, 200},
		'.': []int{200, 300},
		',': []int{0, 300},
	}
	images = make(map[byte]image.Image)
)

func mustSetupEndpoint(address string) *endpoint {
	return &endpoint{address: address}
}

type endpoint struct {
	address string
	srv     *http.Server
}

func (e *endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/counter") {
		e.counter(w, r)
	}
}

func (e *endpoint) serve() {
	mux := http.NewServeMux()
	mux.Handle("/", e)

	srv := &http.Server{
		Addr:    e.address,
		Handler: mux,
	}
	e.srv = srv

	log.Printf("Serving endpoint at [%v]", e.srv.Addr)
	err := srv.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			log.WithFields(log.Fields{"err": err}).Error("endpoint: stopped working")
		}
	}
}

func (e *endpoint) counter(w http.ResponseWriter, r *http.Request) {
	var url = r.URL.String()
	switch r.Method {
	case "GET":
		m[url]++
		var value = []byte(fmt.Sprintf("%v", m[url]))

		buffer := new(bytes.Buffer)
		if err := png.Encode(buffer, createConcatenatedImage(value)); err != nil {
			log.Println("unable to encode image.")
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
		if _, err := w.Write(buffer.Bytes()); err != nil {
			log.Println("unable to write image.")
		}

	case "DELETE":
		delete(m, url)
	}
}

func init() {
	filename := "numbers.png"
	infile, err := os.Open(filename)
	if err != nil {
		panic(err.Error())
	}
	defer infile.Close()

	src, err := png.Decode(infile)
	if err != nil {
		panic(err.Error())
	}

	for b := range cs {
		images[b] = generateNumberImage(b, src)
	}
}

func generateNumberImage(b byte, i image.Image) image.Image {
	rgba := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{100, 100}})
	draw.Draw(rgba, image.Rectangle{image.Point{0, 0}, image.Point{100, 100}}, i, image.Point{cs[b][0], cs[b][1]}, draw.Over)
	return rgba
}

func createConcatenatedImage(bs []byte) image.Image {
	rgba := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{len(bs) * 100, 100}})
	var x = 0
	for i := range bs {
		draw.Draw(rgba, image.Rectangle{image.Point{x, 0}, image.Point{x + 100, 100}}, images[bs[i]], image.Point{0, 0}, draw.Src)
		x += 100
	}
	return rgba
}
