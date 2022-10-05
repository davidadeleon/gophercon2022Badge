package main

import (
	"image/color"
	"machine"
	"reflect"

	"github.com/davidadeleon/gophercon2022Badge/button"
	"github.com/davidadeleon/gophercon2022Badge/il0373"
	"github.com/davidadeleon/gophercon2022Badge/neopixel"
	"tinygo.org/x/drivers/apds9960"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
	"tinygo.org/x/tinyfont/gophers"
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
	sensor      = apds9960.New(machine.I2C1)
	eController = neopixel.EffectsController{APDS9960: &sensor}
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

	// Setup APDS9960 Sensor
	sensor.Configure(apds9960.Configuration{})

	sensor.EnableGesture()

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

	// Display Namebadge
	//NameBadgeDisplay()

	// Create Effects Controller
	eController.Init(&neoPixelStickController)
	eController.Start()

	// Create Button Manager
	buttonManager := button.ButtonManager{
		Watcher: make(chan string),
	}

	// Create Buttons
	aButton := button.Button{
		Name:   "A",
		Pin:    machine.D11,
		Mode:   machine.PinInputPullup,
		Action: eController.ChangeEffect,
	}

	bButton := button.Button{
		Name:   "B",
		Pin:    machine.D12,
		Mode:   machine.PinInputPullup,
		Action: ChangeColor,
	}

	cButton := button.Button{
		Name:   "C",
		Pin:    machine.D13,
		Mode:   machine.PinInputPullup,
		Action: eController.StopEffects,
	}

	// Register Buttons to Manager
	buttonManager.Register(&aButton, &bButton, &cButton)

	// Start Manager
	buttonManager.Start()
}

func NameBadgeDisplay() {
	// Clear ePaper Display
	println("[GC22_Badge] Clearing Display")
	eDisplay.Fill(il0373.WHITE)

	// Lazy fix reversed image
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
}

func ChangeColor() {
	eController.ColorCh <- struct{}{}
}

func ClearPixelBar() {
	neoPixelStickController.Clear()
}

/*
func EffectRGB(stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		default:
			for
		}
	}
}
*/

func reverseSlice(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
