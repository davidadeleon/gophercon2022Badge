package neopixel

import (
	"bytes"
	"machine"

	"tinygo.org/x/drivers/ws2812"
)

var (
	off    = RGBW{0x00, 0x00, 0x00, 0x00}
	Colors = []RGBW{
		RGBW{0x32, 0x00, 0x00, 0x00}, // Red
		RGBW{0x00, 0x32, 0x00, 0x00}, // Green
		RGBW{0x00, 0x00, 0x32, 0x00}, // Blue
		RGBW{0x00, 0x00, 0x00, 0x32}, // White
		RGBW{0x00, 0x16, 0x16, 0x00}, // Gopher
	}
	ColorIndex = 0
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

func (n *NeoPixelController) Fill(color RGBW) {
	colorBuf := new(bytes.Buffer)
	for i := 0; i < len(n.NeoPixels); i++ {
		colorBuf.Write(color.GBRASlice())
	}
	n.Controller.Write(colorBuf.Bytes())
}

func (p *NeoPixel) SetColor(r, g, b, w byte) {
	p.Color = RGBW{r, g, b, w}
}

func (c *RGBW) GBRASlice() []byte {
	return []byte{c.G, c.R, c.B, c.W}
}
