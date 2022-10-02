package main

import (
	"image/color"
	"machine"
	"reflect"
	"time"

	"github.com/davidadeleon/gophercon2022Badge/button"
	"github.com/davidadeleon/gophercon2022Badge/il0373"
	"github.com/davidadeleon/gophercon2022Badge/neopixel"
	"tinygo.org/x/drivers/apds9960"
)

const (
	width  = 128
	height = 296
)

var (
	eDisplay                *il0373.Device
	neoPixelStickController = neopixel.NeoPixelController{
		Pin: machine.GPIO26,
	}

	// RGBA Colors for Text Printing
	//white := color.RGBA{0, 0, 0, 255}
	red   = color.RGBA{255, 0, 0, 255}
	black = color.RGBA{255, 255, 255, 255}
)

func init() {
	// Init SPI0
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 1000000,
		Mode:      0,
	})

	// Init I2C1
	machine.I2C1.Configure(machine.I2CConfig{
		SCL:       machine.SCL_PIN,
		SDA:       machine.SDA_PIN,
		Frequency: machine.TWI_FREQ_400KHZ,
	})
}

func main() {
	println("[GC22_Badge] Starting!")

	// Setup NeoPixel Stick and fill witn inital color
	neoPixelStickController.Init(8)
	neoPixelStickController.Fill(neopixel.Colors[neopixel.ColorIndex])

	// Setup APDS9960 Sensor
	sensor := apds9960.New(machine.I2C1)
	sensor.Configure(apds9960.Configuration{})

	sensor.EnableColor()

	/*
		// Setup IL0373 ePaper Display
		println("[GC22_Badge] Initializing display...")
		eDisplay = il0373.New(
			width,
			height,
			machine.SPI0,  // SPI
			machine.D9,    // CS Pin
			machine.D10,   // DC Pin
			machine.NoPin, // SRAM CS Pin
			machine.NoPin, // RST Pin
			machine.NoPin, // BUSY Pin
		)
		eDisplay.Initialize()
		eDisplay.SetRotation(0)
		println("[GC22_Badge] Initialized!")

		// Clear ePaper Display
		println("[GC22_Badge] Clearing Display")
		eDisplay.Fill(il0373.WHITE)

		// Fix reversed image
		for _, row := range qr_code_github_buf {
			reverseSlice(row)
		}

		// Display
		println("[GC22_Badge] Display...")
		eDisplay.DisplayImage(width, height, qr_code_github_buf)
		tinyfont.WriteLineRotated(eDisplay, &freemono.Bold18pt7b, 35, 290, "David", black, tinyfont.ROTATION_270)
		tinyfont.WriteLineRotated(eDisplay, &freemono.Bold18pt7b, 65, 290, "De Leon", black, tinyfont.ROTATION_270)
		tinyfont.WriteLineRotated(eDisplay, &gophers.Regular58pt, 125, 290, "E", black, tinyfont.ROTATION_270)
		tinyfont.WriteLineRotated(eDisplay, &freemono.Regular9pt7b, 105, 245, "Built using", red, tinyfont.ROTATION_270)
		tinyfont.WriteLineRotated(eDisplay, &freemono.Bold9pt7b, 120, 245, "TinyGo!", red, tinyfont.ROTATION_270)
		eDisplay.Display()
		eDisplay.PowerDown()

		println("[GC22_Badge] Done!")

	*/

	// Create Button Manager
	buttonManager := button.ButtonManager{
		Watcher: make(chan string),
	}

	// Create Buttons
	aButton := button.Button{
		Name:   "A",
		Pin:    machine.D11,
		Mode:   machine.PinInputPullup,
		Action: DumbGlow,
	}

	bButton := button.Button{
		Name:   "B",
		Pin:    machine.D12,
		Mode:   machine.PinInputPullup,
		Action: GoColors,
	}

	cButton := button.Button{
		Name:   "C",
		Pin:    machine.D13,
		Mode:   machine.PinInputPullup,
		Action: CycleColors,
	}

	// Register Buttons to Manager
	buttonManager.Register(&aButton, &bButton, &cButton)

	// Start Manager
	buttonManager.Start()
}

func ClearPixelBar(stopCh <-chan struct{}) {
	neoPixelStickController.Clear()
}

// Gopher Blue!
func GoColors(stopCh <-chan struct{}) {
	select {
	case <-stopCh:
		return
	default:
		neoPixelStickController.Fill(neopixel.RGBW{0x00, 0x16, 0x16, 0x00})
	}
}

func CycleColors(stopCh <-chan struct{}) {
	select {
	case <-stopCh:
		return
	default:
		neopixel.ColorIndex++
		if neopixel.ColorIndex >= len(neopixel.Colors) {
			// Loop back around
			neopixel.ColorIndex = 0
		}
		neoPixelStickController.Fill(neopixel.Colors[neopixel.ColorIndex])
	}
}

func showGC22GH() {
	println("[GC22_Badge] B func")
	eDisplay.Fill(il0373.WHITE)
	eDisplay.Display()
	eDisplay.PowerDown()
}

func reverseSlice(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func DumbGlow(stopCh <-chan struct{}) {
	currStep := 0
	steps := 64
	brighten := true
	color := neopixel.RGBW{
		R: 0x00,
		G: 0xFF,
		B: 0xFF,
		W: 0x00,
	}
	for {
		select {
		case <-stopCh:
			return
		default:
			if brighten {
				currStep++
				println("G:", byte((int(color.G)/steps)*currStep), "B:", byte((int(color.B)/steps)*currStep))
				neoPixelStickController.Fill(neopixel.RGBW{
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
				println("G:", byte((int(color.G)/steps)*currStep), "B:", byte((int(color.B)/steps)*currStep))
				neoPixelStickController.Fill(neopixel.RGBW{
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
