// Adapted From:
//	- https://github.com/adafruit/Adafruit_CircuitPython_framebuf/blob/main/adafruit_framebuf.py

package framebuffer

import (
	"fmt"
)

// TODO: Implement other Formats

// Type FrameBuffer implements a framebuffer using the MHMSBFormat
type FrameBuffer struct {
	Buf      *[]byte
	Width    int
	Height   int
	Stride   int
	Rotation int // Can only be one of (0, 1, 2, 3)
}

// Func SetRotation sets the rotation of the framebuffer
func (f *FrameBuffer) SetRotation(val int) error {
	if val > 3 {
		return fmt.Errorf("invalid rotation setting")
	}
	f.Rotation = val
	return nil
}

func getPixel(fb *FrameBuffer, x, y int) byte {
	index := (y*fb.Stride + x) / 8
	offset := 7 - x&0x07
	return ((*fb.Buf)[index] >> offset) & 0x01
}

func setPixel(fb *FrameBuffer, x, y, color int) {
	index := (y*fb.Stride + x) / 8
	offset := 7 - (x & 0x07)
	colorBit := uint8(0)
	if color != 0 {
		colorBit = 1
	}
	(*fb.Buf)[index] = ((*fb.Buf)[index] & ^(0x01 << offset)) | (colorBit << offset)
}

// Func Pixel will get the value of a pixel if you don't pass a color
// if you pass a color, Pixel will set the pixel to the provided color
func (f *FrameBuffer) Pixel(x, y int, color ...int) byte {
	switch f.Rotation {
	case 1:
		x, y = y, x
		x = f.Width - x - 1
	case 2:
		x = f.Width - x - 1
		y = f.Height - y - 1
	case 3:
		x, y = y, x
		y = f.Height - y - 1
	}

	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		return 0
	}

	l := len(color)
	switch l {
	case 0:
		return getPixel(f, x, y)
	case 1:
		setPixel(f, x, y, color[0])
		return 0
	default:
		return 0
	}
}

func (f *FrameBuffer) Clear() {
	for i := range *f.Buf {
		(*f.Buf)[i] = 0x00
	}
}

func (f *FrameBuffer) Fill(color uint8) {
	for i := 0; i < len((*f.Buf)); i++ {
		(*f.Buf)[i] = color
	}
}

func (f *FrameBuffer) FillRect(x, y, width, height int, color uint8) {
	for _x := x; _x < (x + width); _x++ {
		offset := 7 - (_x & 0x07)
		for _y := y; _y < (y + height); _y++ {
			index := (_y*f.Stride + _x) / 8
			(*f.Buf)[index] = ((*f.Buf)[index] & ^(0x01 << offset)) | (color << offset)
		}
	}
}

func (f *FrameBuffer) Rect(x, y, width, height int, color byte, fill bool) {
	switch f.Rotation {
	case 1:
		x, y = y, x
		width, height = height, width
		x = f.Width - x - width
	case 2:
		x = f.Width - x - width
		y = f.Height - y - height
	case 3:
		x, y = y, x
		width, height = height, width
		y = f.Height - y - height
	}

	if width < 1 ||
		height < 1 ||
		(x+width) <= 0 ||
		(y+height) <= 0 ||
		y >= f.Height ||
		x >= f.Width {
		return
	}

	x_end := min((f.Width - 1), (x + width - 1))
	y_end := min((f.Height - 1), (y + height - 1))
	x = max(x, 0)
	y = max(y, 0)

	if fill {
		f.FillRect(x, y, x_end-x+1, y_end-y+1, color)
		return
	}
	f.FillRect(x, y, (x_end - x + 1), 1, color)
	f.FillRect(x, y, 1, (y_end - y + 1), color)
	f.FillRect(x, y_end, (x_end - x + 1), 1, color)
	f.FillRect(x_end, y, 1, (y_end - y + 1), color)
}

func (f *FrameBuffer) Line(x0, y0, x1, y1, color int) {
	//Bresenham's line algorithm
	var x, y, sx, sy int

	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	x = x0
	y = y0

	if x0 > x1 {
		sx = -1
	} else {
		sx = 1
	}

	if y0 > y1 {
		sy = -1
	} else {
		sy = 1
	}

	if dx > dy {
		e := dx / 2.0
		for x != x1 {
			f.Pixel(x, y, color)
			e -= dy
			if e < 0 {
				y += sy
				e += dx
			}
			x += sx
		}
	} else {
		e := dy / 2.0
		for y != y1 {
			f.Pixel(x, y, color)
			e -= dx
			if e < 0 {
				x += sx
				e += dy
			}
			y += sy
		}
	}

	f.Pixel(x, y, color)
}

func (f *FrameBuffer) HLine(x, y, width int, color byte) {
	f.Rect(x, y, width, 1, color, true)
}

func (f *FrameBuffer) VLine(x, y, height int, color byte) {
	f.Rect(x, y, 1, height, color, true)
}

func (f *FrameBuffer) Circle(centerx, centery, radius, color int) {
	var x, y, dx, dy, e int
	x = radius - 1
	y = 0
	dx = 1
	dy = 1
	e = dx - (radius << 1)
	for x >= y {
		f.Pixel(centerx+x, centery+y, color)
		f.Pixel(centerx+y, centery+x, color)
		f.Pixel(centerx-y, centery+x, color)
		f.Pixel(centerx-x, centery+y, color)
		f.Pixel(centerx-x, centery-y, color)
		f.Pixel(centerx-y, centery-x, color)
		f.Pixel(centerx+y, centery-x, color)
		f.Pixel(centerx+x, centery-y, color)
		if e <= 0 {
			y += 1
			e += dy
			dy += 2
		}
		if e > 0 {
			x -= 1
			dx += 2
			e += dx - (radius << 1)
		}
	}
}

// Func Scroll shifts the framebuffer in x and y direction
func (f *FrameBuffer) Scroll(dx, dy int) {
	var shiftx, xend, yend, dtx, dty, x, y int
	if dx < 0 {
		shiftx = 0
		xend = f.Width + dx
		dtx = 1
	} else {
		shiftx = f.Width - 1
		xend = dx - 1
		dtx = -1
	}
	if dy < 0 {
		y = 0
		yend = f.Height + dx
		dty = 1
	} else {
		y = f.Height - 1
		yend = dy - 1
		dty = -1
	}
	for y != yend {
		x = shiftx
		for x != xend {
			setPixel(f, x, y, int(getPixel(f, (x-dx), (y-dy))))
			x += dtx
		}
		y += dty
	}
}

// Rough implementation just for B/W
func (f *FrameBuffer) DisplayImage(imageWidth, imageHeight int, image []byte) {
	width := f.Width
	height := f.Height
	if f.Rotation != 0 {
		width, height = height, width
	}
	if imageWidth != width || imageHeight != height {
		println("error displaying image: image must be same dimensions as display")
	}

	// Clear Buffer
	f.Fill(0)
	/*
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				pixel = image[x, y]
				f.Pixel(x, y, 1)
			}
		}
	*/
}

/*
func (f *FrameBuffer) Text(text string, x,y int, color uint8, font_name string, size int) {
	frame_width := f.Width
	frame_height := f.Height
	if f.Rotation == 1 || f.Rotation == 3{
		frame_width, frame_height = frame_height, frame_width
	}

	for chunk
}
*/

// Min returns the smallest of two parameters
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of the two parameters
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
