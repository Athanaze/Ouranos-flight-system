package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"encoding/binary"

	"github.com/gorilla/websocket"
)

type Packet struct {
	Category category
	Content  string
}

type Message struct {
	Content []byte
}

var (
	mu             sync.Mutex
	messagesToSend []Message
)

type category byte
type tickStamp uint32

const (
	C_ORIENTATION_POSITION     category = 0
	C_FLIGHT_MODE     category = 2
	C_BATTERY_VOLTAGE category = 3
	C_AIRSPEED        category = 4
)

type Orientation struct {
	Tick  tickStamp
	Roll  float32
	Pitch float32
	Yaw   float32
}

type Position struct {
	Tick tickStamp,
	Altitude float32,
	GPSLat float32,
	GPSLng float32,
	Laser float32
}

func main() {

	////////////////////////////////////////////////////////////////////////
	// file server part
	fs := http.FileServer(
		//http.Dir("C:\\Users\\Sacha\\Documents\\web_ground_station\\web_ground_station\\web_instruments\\web")
		http.Dir("web")
	)
	http.Handle("/", fs)

	/////////////////////////////////////////////////////////////////////////
	// API Endpoints
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ouais")
	})
	http.HandleFunc("/ws/", ws)
	go serialListen()
	log.Println("Listening on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

var upgrader = websocket.Upgrader{} // use default options

func ws(w http.ResponseWriter, r *http.Request) {
	
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	//defer c.Close()
	for {
		
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read error :", err)
			break
		}
		log.Printf(string(message))
		for {
			if len(messagesToSend) != 0 {
				mu.Lock()
				m := messagesToSend[0]
				messagesToSend = messagesToSend[1:]
				mu.Unlock()
				err = c.WriteMessage(websocket.BinaryMessage, m.Content)
				if err != nil {
					log.Println("write:", err)
					break
				}
			}
		}
	}
}

func Float64bytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}
func Float32bytes(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}

func (o *Orientation) toBytes() []byte {
	headerSlice := append([]byte{byte(C_ORIENTATION_POSITION)}, o.Tick.toBytes()...)
	a := append(Float32bytes(o.Pitch), Float32bytes(o.Yaw)...)
	dataSlice := append(
		Float32bytes(o.Roll),
		a...,
	)

	b := append(Float32bytes(p.Altitude), Float32bytes(o.GPSLat)...)
	dataSliceB := append(
		Float32bytes(p.GPSLng),
		b...,
	)

	dataSlice2 := append(Float32bytes(p.Laser), dataSliceB...)

	return append(headerSlice, dataSlice2...)
}

func (p *Position) toBytes() []byte {
	headerSlice := append([]byte{byte(C_POSITION)}, p.Tick.toBytes()...)
	

	return append(headerSlice, dataSlice2...)
}

func (t tickStamp) toBytes() []byte {
	a := make([]byte, 4)
	binary.LittleEndian.PutUint32(a, uint32(t))
	return a
}

func serialListen() {
	var tick tickStamp = 0
	for {
		tick++
		time.Sleep(100 * time.Millisecond)
		mu.Lock()
		// We assume the packet we receive has been successfully read as a "Orientation" packet
		o := OrientationPosition{
			Tick: tick,
			Roll: 5.2 * rand.Float32(),
			Pitch: 3.421 * rand.Float32(),
			Yaw: 19.56 * rand.Float32(),
			Altitude: 34 * rand.Float32(),
			GPSLat: 43.0232254,
			GPSLng: 7.4345233,
			Laser: 99.123
		}
		mOrientationPosition := Message{Content: o.toBytes()}
		
		messagesToSend = append(messagesToSend, mOrientationPosition)
		
		mu.Unlock()
		log.Println("MessagesToSend length : ")
		log.Println(len(messagesToSend))
	}
}
