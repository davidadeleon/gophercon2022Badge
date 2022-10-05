package neopixel

import (
	"bytes"
	"machine"
	"time"

	"tinygo.org/x/drivers/apds9960"
	"tinygo.org/x/drivers/ws2812"
)

var (
	off    = RGBW{0x00, 0x00, 0x00, 0x00}
	Colors = []RGBW{
		RGBW{0x64, 0x00, 0x00, 0x00}, // Red
		RGBW{0x00, 0x64, 0x00, 0x00}, // Green
		RGBW{0x00, 0x00, 0x64, 0x00}, // Blue
		RGBW{0x00, 0x00, 0x00, 0x64}, // White
		RGBW{0x00, 0x64, 0x64, 0x00}, // Gopher
		RGBW{0x64, 0x64, 0x00, 0x00}, // Yellow
		RGBW{0x64, 0x32, 0x00, 0x00}, // Orange
	}
	ColorIndex = 0
	PixelIndex = 0
)

type NeoPixelController struct {
	Pin        machine.Pin
	Controller ws2812.Device
	NeoPixels  []*NeoPixel
	Enabled    bool
}

type NeoPixel struct {
	Color RGBW
}

type RGBW struct {
	R byte
	G byte
	B byte
	W byte
}

func (n *NeoPixelController) Init(pixels int) {
	n.Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	n.Controller = ws2812.New(n.Pin)
	for i := 0; i < pixels; i++ {
		n.NeoPixels = append(n.NeoPixels, &NeoPixel{RGBW{0x00, 0x00, 0x00, 0x00}})
	}
}

func (n *NeoPixelController) Show() {
	n.Clear()
	colorBuf := new(bytes.Buffer)
	for _, pixel := range n.NeoPixels {
		colorBuf.Write(pixel.Color.GBRASlice())
	}
	n.Controller.Write(colorBuf.Bytes())
}

func (n *NeoPixelController) Clear() {
	colorBuf := make([]byte, (len(n.NeoPixels)))
	for i := 0; i < len(n.NeoPixels); i++ {
		colorBuf = append(colorBuf, off.GBRASlice()...)
	}
	n.Controller.Write(colorBuf)
}

func (n *NeoPixelController) ClearPixels() {
	for _, pixel := range n.NeoPixels {
		pixel.SetColor(off)
	}
	n.Show()
}

func (n *NeoPixelController) Fill(color RGBW) {
	colorBuf := new(bytes.Buffer)
	for i := 0; i < len(n.NeoPixels); i++ {
		colorBuf.Write(color.GBRASlice())
	}
	n.Controller.Write(colorBuf.Bytes())
}

// Func ShowPixel sets a Single Pixel Color
func (n *NeoPixelController) SetPixel(index int, color RGBW) {
	n.NeoPixels[index].SetColor(color)
}

func (n *NeoPixelController) ChangeColor() {
	ColorIndex++
	if ColorIndex >= (len(Colors) - 1) {
		ColorIndex = 0
	}
}

func (p *NeoPixel) SetColor(color RGBW) {
	p.Color = color
}

func (c *RGBW) GBRASlice() []byte {
	return []byte{c.G, c.R, c.B, c.W}
}

// NeoPixel Effects

type EffectsController struct {
	Effects       []func(*chan struct{})
	EffectIndex   int
	msgCh         chan struct{}
	ColorCh       chan struct{}
	stopCh        chan struct{}
	nController   *NeoPixelController
	APDS9960      *apds9960.Device
	EffectRunning bool
}

func (e *EffectsController) Init(n *NeoPixelController) {
	e.nController = n
	e.EffectIndex = 0
	e.Effects = append(e.Effects, e.Static)
	e.Effects = append(e.Effects, e.CycleColors)
	e.Effects = append(e.Effects, e.KnightRider)
	e.Effects = append(e.Effects, e.Glow)

	e.msgCh = make(chan struct{})
}

func (e *EffectsController) Start() {
	if e.APDS9960 != nil {
		e.Effects = append(e.Effects, e.ProximitySense)
		e.Effects = append(e.Effects, e.GestureSense)
	}

	// Kick off the controller
	e.stopCh = make(chan struct{})
	e.ColorCh = make(chan struct{})

	// Start the first Effect
	go e.Effects[e.EffectIndex](&e.stopCh)
	e.EffectRunning = true

	go func() {
		for {
			select {
			case <-e.msgCh:
				if e.EffectRunning {
					println("Sending Stop from controller")
					e.stopCh <- struct{}{}
					e.EffectRunning = false
				}
				e.EffectIndex++
				if e.EffectIndex > (len(e.Effects) - 1) {
					e.EffectIndex = 0
				}
				e.EffectRunning = true
				go e.Effects[e.EffectIndex](&e.stopCh)
			case <-e.ColorCh:
				ColorIndex++
				if ColorIndex > (len(Colors) - 1) {
					ColorIndex = 0
				}
			case <-e.stopCh:
				println("Got Stop at controller")
				e.stopCh <- struct{}{}
				e.EffectRunning = false
			}
		}
	}()
}

func (e *EffectsController) ChangeEffect() {
	e.msgCh <- struct{}{}
}

func (e *EffectsController) StopEffects() {
	e.stopCh <- struct{}{}
	e.nController.ClearPixels()
}

func (e *EffectsController) Static(stopCh *chan struct{}) {
	println("Static")
	for {
		select {
		case <-*stopCh:
			return
		default:
			e.nController.Fill(Colors[ColorIndex])
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (e *EffectsController) CycleColors(stopCh *chan struct{}) {
	println("Cycle Colors")
	for {
		select {
		case <-*stopCh:
			return
		default:
			ColorIndex++
			if ColorIndex >= len(Colors) {
				// Loop back around
				ColorIndex = 0
			}
			e.nController.Fill(Colors[ColorIndex])
			time.Sleep(1 * time.Second)
		}
	}
}

func (e *EffectsController) KnightRider(stopCh *chan struct{}) {
	println("Hasselhoff")
	forward := true
	for {
		select {
		case <-*stopCh:
			println("Stopping Hasselhoff")
			return
		default:
			if forward {
				e.nController.ClearPixels()
				PixelIndex++
				if PixelIndex >= len(e.nController.NeoPixels)-1 {
					forward = false
					// Handle the event that we pause right on the last index and restart
					PixelIndex = len(e.nController.NeoPixels) - 1
				}
				e.nController.SetPixel(PixelIndex, Colors[ColorIndex])
				e.nController.Show()
				time.Sleep(50 * time.Millisecond)
			} else {
				e.nController.ClearPixels()
				PixelIndex--
				if PixelIndex <= 0 {
					forward = true
					// Handle the event that we pause right on the last index and restart
					PixelIndex = 0
				}
				e.nController.SetPixel(PixelIndex, Colors[ColorIndex])
				e.nController.Show()
				time.Sleep(50 * time.Millisecond)
			}
		}
	}
}

func (e *EffectsController) Glow(stopCh *chan struct{}) {
	println("Glow")
	currStep := 0
	steps := 64
	brighten := true
	for {
		color := Colors[ColorIndex]
		select {
		case <-*stopCh:
			println("Stopping Glow")
			return
		default:
			if brighten {
				currStep++
				e.nController.Fill(RGBW{
					R: byte((int(color.R) / steps) * currStep),
					G: byte((int(color.G) / steps) * currStep),
					B: byte((int(color.B) / steps) * currStep),
					W: byte((int(color.W) / steps) * currStep),
				})
				time.Sleep(10 * time.Millisecond)
				if currStep == steps {
					brighten = false
				}

			} else {
				currStep--
				e.nController.Fill(RGBW{
					R: byte((int(color.R) / steps) * currStep),
					G: byte((int(color.G) / steps) * currStep),
					B: byte((int(color.B) / steps) * currStep),
					W: byte((int(color.W) / steps) * currStep),
				})
				time.Sleep(10 * time.Millisecond)
				if currStep == 0 {
					brighten = true
				}
			}
		}
	}
}

func (e *EffectsController) sweep(direction string) {
	if direction == "left" {
		for i := 0; i < len(e.nController.NeoPixels); i++ {
			e.nController.ClearPixels()
			e.nController.SetPixel(i, Colors[ColorIndex])
			e.nController.Show()
			time.Sleep(50 * time.Millisecond)
		}
		e.nController.ClearPixels()
		e.nController.Show()
	} else {
		for i := len(e.nController.NeoPixels) - 1; i >= 0; i-- {
			e.nController.ClearPixels()
			e.nController.SetPixel(i, Colors[ColorIndex])
			e.nController.Show()
			time.Sleep(50 * time.Millisecond)
		}
		e.nController.ClearPixels()
		e.nController.Show()
	}
}

func (e *EffectsController) ProximitySense(stopCh *chan struct{}) {
	println("Proximity")
	e.APDS9960.EnableProximity()
	for {
		select {
		case <-*stopCh:
			return
		default:
			if e.APDS9960.ProximityAvailable() {
				p := e.APDS9960.ReadProximity()
				e.nController.ClearPixels()
				var endIndex int
				switch {
				case p == 255:
					// Too Far to Measure
					continue
				case p > 224:
					endIndex = 7
				case p > 192:
					endIndex = 6
				case p > 160:
					endIndex = 5
				case p > 128:
					endIndex = 4
				case p > 96:
					endIndex = 3
				case p > 64:
					endIndex = 2
				case p > 32:
					endIndex = 1
				case p >= 0:
					endIndex = 0
				}
				for i := len(e.nController.NeoPixels) - 1; i >= endIndex; i-- {
					e.nController.SetPixel(i, Colors[ColorIndex])
				}
				e.nController.Show()
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

func (e *EffectsController) GestureSense(stopCh *chan struct{}) {
	println("Gesture")
	e.APDS9960.EnableGesture()
	e.APDS9960.Setsensitivity(100)
	e.APDS9960.SetGains(1, 4, 4)
	e.APDS9960.LEDBoost(300)
	e.nController.ClearPixels()
	var gesture int32
	gesture = 0
	for {
		select {
		case <-*stopCh:
			return
		default:
			if e.APDS9960.GestureAvailable() {
				gesture = e.APDS9960.ReadGesture()
				switch gesture {
				case apds9960.GESTURE_UP:
					println("Up")
					// Because the sensor is rotated
					e.sweep("right")
				case apds9960.GESTURE_DOWN:
					println("Down")
					// Because the sensor is rotated
					e.sweep("left")
				case apds9960.GESTURE_LEFT:
					println("Left")
					e.sweep("right")
				case apds9960.GESTURE_RIGHT:
					println("Right")
					e.sweep("left")
				}
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
}
