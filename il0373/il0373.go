package il0373

import (
	"bytes"
	"encoding/binary"
	"image/color"
	"machine"
	"time"

	"github.com/davidadeleon/gophercon2022Badge/framebuffer"
	"tinygo.org/x/drivers"
)

const (
	BLACK int = iota
	WHITE
	INVERSE
	RED
	DARK
	LIGHT
)

type Device struct {
	SPI         drivers.SPI
	CS_PIN      machine.Pin
	DC_PIN      machine.Pin
	SRAM_CS_PIN machine.Pin
	RST_PIN     machine.Pin
	BUSY_PIN    machine.Pin

	Width        int
	Height       int
	Rotation     int
	buffer1_size int
	buffer2_size int

	buffer1 []byte
	buffer2 []byte

	framebuf1        *framebuffer.FrameBuffer
	framebuf2        *framebuffer.FrameBuffer
	blackFrameBuffer *framebuffer.FrameBuffer
	colorFrameBuffer *framebuffer.FrameBuffer
	blackInverted    bool
	colorInverted    bool

	spiBuf       []byte
	singleByteTx bool

	_buf []byte
}

func New(width, height int, bus drivers.SPI, csPin, dcPin, sramCSPin, rstPin, busyPin machine.Pin) *Device {
	return &Device{
		Width:       width,
		Height:      height,
		SPI:         bus,
		CS_PIN:      csPin,
		DC_PIN:      dcPin,
		SRAM_CS_PIN: sramCSPin,
		RST_PIN:     rstPin,
		BUSY_PIN:    busyPin,
	}
}

func (d *Device) Initialize() {
	// Setup reset pin if provided
	if d.RST_PIN != machine.NoPin {
		d.RST_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}

	// Setup busy pin if provided
	if d.BUSY_PIN != machine.NoPin {
		d.BUSY_PIN.Configure(machine.PinConfig{Mode: machine.PinInput})
	}

	// Setup DC Pin (required)
	d.DC_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.DC_PIN.Low()

	// Setup CS Pin (required)
	d.CS_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.CS_PIN.High()

	d.spiBuf = make([]byte, 1)
	d.singleByteTx = false

	// TODO: Implement SRAM

	d._buf = make([]byte, 3)

	d.buffer1_size = int(d.Width * d.Height / 8)
	d.buffer2_size = d.buffer1_size

	d.buffer1 = make([]byte, (d.Width*d.Height)/8)
	d.buffer2 = make([]byte, (d.Width*d.Height)/8)

	d.framebuf1 = &framebuffer.FrameBuffer{
		Buf:      &d.buffer1,
		Width:    d.Width,
		Height:   d.Height,
		Stride:   d.Width, // No Calc for this so default to width
		Rotation: d.Rotation,
	}

	d.framebuf2 = &framebuffer.FrameBuffer{
		Buf:      &d.buffer2,
		Width:    d.Width,
		Height:   d.Height,
		Stride:   d.Width, // No Calc for this so default to width
		Rotation: d.Rotation,
	}

	d.SetBlackBuffer(0, true)
	d.SetColorBuffer(1, true)
	d.HardwareReset()
}

//=============================================================\\
//                   il0373 Specific Functions                 \\
//=============================================================\\

func (d *Device) Begin(reset bool) {
	if reset {
		d.HardwareReset()
	}
	d.PowerDown()
}

func (d *Device) BusyWait() {
	if d.BUSY_PIN != machine.NoPin {
		for d.BUSY_PIN.Get() {
			time.Sleep(10 * time.Millisecond)
		}
		return
	}
	time.Sleep(500 * time.Millisecond)
}

func (d *Device) PowerUp() error {
	d.HardwareReset()
	d.BusyWait()

	d.command(IL0373_POWER_SETTING, []byte{0x03, 0x00, 0x2B, 0x2B, 0x09}, true)
	d.command(IL0373_BOOSTER_SOFT_START, []byte{0x17, 0x17, 0x17}, true)
	d.command(IL0373_POWER_ON, nil, true)

	d.BusyWait()
	time.Sleep(200 * time.Millisecond)

	d.command(IL0373_PANEL_SETTING, []byte{0xCF}, true)
	d.command(IL0373_CDI, []byte{0x37}, true)
	d.command(IL0373_PLL, []byte{0x29}, true)

	_b1 := byte(d.Width & 0xFF)
	_b2 := byte((d.Height >> 8) & 0xFF)
	_b3 := byte(d.Height & 0xFF)
	d.command(IL0373_RESOLUTION, []byte{_b1, _b2, _b3}, true)
	d.command(IL0373_VCM_DC_SETTING, []byte{0x0A}, true)
	time.Sleep(20 * time.Millisecond)

	return nil
}

func (d *Device) PowerDown() {
	d.command(IL0373_CDI, []byte{0x17}, true)
	d.command(IL0373_VCM_DC_SETTING, []byte{0x00}, true)
	d.command(IL0373_POWER_OFF, nil, true)
}

func (d *Device) Update() {
	d.command(IL0373_DISPLAY_REFRESH, nil, true)
	time.Sleep(100 * time.Millisecond)
	d.BusyWait()
	println("[IL0373] device waiting...")
	time.Sleep(15 * time.Second)
	println("[IL0373] done waiting!")
}

func (d *Device) WriteRam(index int) byte {
	switch index {
	case 0:
		return d.command(IL0373_DTM1, nil, false)
	case 1:
		return d.command(IL0373_DTM2, nil, false)
	default:
		return 0
	}
}

//=============================================================\\
//                     Generic EPD Functions                   \\
//=============================================================\\

func (d *Device) Display() error {
	d.PowerUp()

	d.WriteRam(0)
	d.DC_PIN.High()

	for i := 0; i < d.buffer1_size; i++ {
		d.SPITransfer(d.buffer1[i])
	}

	d.CS_PIN.High()
	time.Sleep(20 * time.Millisecond)

	if d.buffer2_size != 0 {
		d.WriteRam(1)

		time.Sleep(10 * time.Millisecond)
		d.DC_PIN.High()

		for i := 0; i < d.buffer2_size; i++ {
			d.SPITransfer(d.buffer2[i])
		}

		d.CS_PIN.High()
	}

	d.Update()
	return nil
}

func (d *Device) HardwareReset() {
	// If we assigned a reset pin, do hardware reset
	if d.RST_PIN != machine.NoPin {
		d.RST_PIN.Low()
		time.Sleep(100 * time.Millisecond)
		d.RST_PIN.High()
		time.Sleep(100 * time.Millisecond)
	}
}

func (d *Device) command(cmd byte, data []byte, end bool) byte {
	d.CS_PIN.High()
	d.DC_PIN.Low()
	d.CS_PIN.Low()

	ret := d.SPITransfer(cmd)

	if data != nil {
		d.DC_PIN.High()
		for i := 0; i < len(data); i++ {
			d.SPITransfer(data[i])
		}
	}
	if end {
		d.CS_PIN.High()
	}
	return ret
}

func (d *Device) SPITransfer(databyte byte) byte {
	d.spiBuf[0] = databyte
	err := d.SPI.Tx(d.spiBuf, d.spiBuf)
	if err != nil {
		println(err)
	}
	return d.spiBuf[0]
}

func (d *Device) SetBlackBuffer(index int, inverted bool) {
	switch index {
	case 0:
		d.blackFrameBuffer = d.framebuf1
	case 1:
		d.blackFrameBuffer = d.framebuf2
	default:
		println("Buffer index must be 0 or 1")
	}
	d.blackInverted = inverted
}

func (d *Device) SetColorBuffer(index int, inverted bool) {
	switch index {
	case 0:
		d.colorFrameBuffer = d.framebuf1
	case 1:
		d.colorFrameBuffer = d.framebuf2
	default:
		println("Buffer index must be 0 or 1")
	}
	d.colorInverted = inverted
}

func ColorDup(f func(), color byte, args ...interface{}) {

}

func (d *Device) SetRotation(val int) {
	d.blackFrameBuffer.SetRotation(val)
	d.colorFrameBuffer.SetRotation(val)
}

func (d *Device) Clear() {
	d.blackFrameBuffer.Clear()
	d.colorFrameBuffer.Clear()
}

func (d *Device) Fill(color int) {
	isRed := color == RED
	redInverted := 0
	if isRed != d.colorInverted {
		redInverted = 1
	}

	red_fill := uint8(redInverted) * 0xFF

	isBlack := color == BLACK
	blackInverted := 0
	if isBlack != d.blackInverted {
		blackInverted = 1
	}

	black_fill := uint8(blackInverted) * 0xFF
	d.blackFrameBuffer.Fill(black_fill)
	d.colorFrameBuffer.Fill(red_fill)
}

func (d *Device) FillRect(x, y, width, height int, color int) {
	// Monochrome
	if d.blackFrameBuffer == d.colorFrameBuffer {
		d.blackFrameBuffer.FillRect(x, y, width, height, byte(color))
		return
	}
	d.blackFrameBuffer.FillRect(x, y, width, height, byte(color))
	d.colorFrameBuffer.FillRect(x, y, width, height, byte(color))
}

func (d *Device) Pixel(x, y int, color int) {
	// Monochrome
	if d.blackFrameBuffer == d.colorFrameBuffer {
		d.blackFrameBuffer.Pixel(x, y, color)
		return
	}

	if (color == BLACK) != d.blackInverted {
		d.blackFrameBuffer.Pixel(x, y, 1)
	} else {
		d.blackFrameBuffer.Pixel(x, y, 0)
	}

	if (color == RED) != d.colorInverted {
		d.colorFrameBuffer.Pixel(x, y, 1)
	} else {
		d.colorFrameBuffer.Pixel(x, y, 0)
	}
}

// Satisfy Displayer Interface to use TinyFonts

// Func Size returns size of display
func (d *Device) Size() (int16, int16) {
	return int16(d.Width), int16(d.Height)
}

// SetPixel modifies the internal buffer.
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	// RED
	if c.R == 255 && c.G == 0 && c.B == 0 {
		d.Pixel(int(x), int(y), RED)
		return
	}
	// White
	if c.R == 0 && c.G == 0 && c.B == 0 {
		d.Pixel(int(x), int(y), WHITE)
		return
	}
	// Black
	if c.R == 255 && c.G == 255 && c.B == 255 {
		d.Pixel(int(x), int(y), BLACK)
		return
	}
}

func WriteBuf(buf *bytes.Buffer, vars ...interface{}) error {
	for elem := range vars {
		err := binary.Write(buf, binary.LittleEndian, elem)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) DisplayImage(imageWidth, imageHeight int, image [][]byte) {
	width := d.Width
	height := d.Height
	if d.Rotation != 0 {
		width, height = height, width
	}
	if imageWidth != width || imageHeight != height {
		println("error displaying image: image must be same dimensions as display")
		return
	}

	// Clear Buffer
	d.Fill(1)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			d.blackFrameBuffer.Pixel(x, y, int(image[x][y]))
		}
	}

}

func (d *Device) LoadImage()
