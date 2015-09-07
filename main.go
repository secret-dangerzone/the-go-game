package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"runtime"

	"github.com/veandco/go-sdl2/sdl"
	"golang.org/x/net/websocket"
	"gopkg.in/vmihailenco/msgpack.v2"
)

var (
	winTitle  = "The Game"
	winWidth  = 1000
	winHeight = 600
)

type Ship struct {
	Position V2
	V        float64
	Rotation float64
	Size     float64
	Doge     bool
}

// Points returns []sdl.Point for the ship body (and doge)
func (s Ship) Points() []sdl.Point {

	// The ship triangle
	points := V2s{
		{0, -s.Size / 2},
		{s.Size / 2, s.Size / 2},
		{-s.Size / 2, s.Size / 2},
		{0, -s.Size / 2},
	}

	// Add a square if the ship is the doge
	if s.Doge {
		points = points.Merge(V2s{
			{-4, -4},
			{-4, 4},
			{4, 4},
			{4, -4},
			{-4, -4},
		})
	}

	return points.Rotate(-s.Rotation).ToPointsOffset(s.Position)
}

type join struct {
	Join interface{} `msgpack:"join"`
}

func joinMessage(name string) []byte {
	n := struct {
		Name string `msgpack:"name"`
	}{
		Name: name,
	}
	m := join{Join: n}

	b, _ := msgpack.Marshal(m)
	return b
}

type InputState struct {
	Left   bool `msgpack:"left"`
	Right  bool `msgpack:"right"`
	Thrust bool `msgpack:"thrust"`
}

func (i InputState) Message() []byte {
	a := struct {
		InputState InputState `msgpack:"inputState"`
	}{
		InputState: i,
	}

	b, _ := msgpack.Marshal(a)
	return b
}

func (i InputState) IsSameAs(o InputState) bool {
	return i.Left == o.Left && i.Right == o.Right && i.Thrust == o.Thrust
}

func (i *InputState) Update(o InputState) {
	i.Left = o.Left
	i.Right = o.Right
	i.Thrust = o.Thrust
}

func run(serverStr, name string) int {
	var window *sdl.Window
	var renderer *sdl.Renderer
	var running bool
	var event sdl.Event
	var inputState InputState

	url := "ws://" + serverStr
	origin := "http://" + serverStr
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer window.Destroy()

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}
	defer renderer.Destroy()

	renderer.Clear()

	websocket.Message.Send(ws, joinMessage(name))

	running = true

	var data map[string]map[string][]map[string]interface{}
	decoder := msgpack.NewDecoder(ws)

	renderer.SetDrawColor(1, 1, 1, 255)
	renderer.Clear()

	for running {
		renderer.SetDrawColor(1, 1, 1, 255)
		renderer.Clear()
		newInputState := inputState

		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false

			case *sdl.KeyDownEvent:
				switch event.(*sdl.KeyDownEvent).Keysym.Sym {
				case sdl.K_ESCAPE:
					running = false
				case sdl.K_LEFT:
					newInputState.Left = true
				case sdl.K_RIGHT:
					newInputState.Right = true
				case sdl.K_UP:
					newInputState.Thrust = true
				}

			case *sdl.KeyUpEvent:
				switch event.(*sdl.KeyUpEvent).Keysym.Sym {
				case sdl.K_LEFT:
					newInputState.Left = false
				case sdl.K_RIGHT:
					newInputState.Right = false
				case sdl.K_UP:
					newInputState.Thrust = false
				}
			}
		}

		if !newInputState.IsSameAs(inputState) {
			websocket.Message.Send(ws, newInputState.Message())
			inputState = newInputState
		}

		decoder.Decode(&data)

		for _, ship := range data["state"]["ships"] {
			colourhex, _ := hex.DecodeString(ship["colour"].(string)[1:7])
			rot, _ := ship["rotation"].(float64)
			s := Ship{
				Position: V2{float64(ship["x"].(int64)), float64(ship["y"].(int64))},
				Size:     30,
				Rotation: rot,
				Doge:     ship["doge"].(bool),
			}
			renderer.SetDrawColor(uint8(colourhex[0]), uint8(colourhex[1]), uint8(colourhex[2]), 255)
			renderer.DrawLines(s.Points())
		}

		renderer.Present()
	}

	return 0
}

func main() {
	runtime.LockOSThread()
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %v [SERVER] [NAME]\n", os.Args[0])
		os.Exit(1)
	}
	os.Exit(run(os.Args[1], os.Args[2]))
}
