package ui

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <stdlib.h>

void DestroyXImage(XImage *image) {
	image->f.destroy_image(image);
}

*/
import "C"
import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"unsafe"
)

var width C.uint
var height C.uint

func Run() {
	display := C.XOpenDisplay(nil)
	if display == nil {
		panic("Unable to open X display")
	}
	defer C.XCloseDisplay(display)

	screen := C.XDefaultScreen(display)

	window := C.XCreateSimpleWindow(
		display,
		C.XRootWindow(display, screen),
		100, 100,
		800, 600,
		1,
		C.XBlackPixel(display, screen),
		C.XWhitePixel(display, screen),
	)

	C.XSelectInput(display, window, C.ExposureMask|C.KeyPressMask|C.StructureNotifyMask)

	C.XMapWindow(display, window)

	for {
		var event C.XEvent
		C.XNextEvent(display, &event)

		switch eventType(event) {
		case C.Expose:
			drawBlue(display, window, screen)
			img, err := loadImageFromEmbed()
			if err == nil {
				drawImageRGBA(display, window, img)
			}
		case C.ConfigureNotify:
			configureEvent := (*C.XConfigureEvent)(unsafe.Pointer(&event))
			width = C.uint(configureEvent.width)
			height = C.uint(configureEvent.height)
		case C.KeyPress:
			return
		}
	}
}

//go:embed image.png
var pngContent []byte

func loadImageFromEmbed() (image.Image, error) {
	img, err := png.Decode(bytes.NewReader(pngContent))
	return img, err
}

func eventType(event C.XEvent) int {
	return int(*(*C.int)(unsafe.Pointer(&event)))
}

func drawImageRGBA(display *C.Display, window C.Window, img image.Image) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	pixels := make([]byte, width*height*4)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			offset := ((y-bounds.Min.Y)*width + (x - bounds.Min.X)) * 4
			r, g, b, a := img.At(x, y).RGBA()

			pixels[offset+0] = byte(b >> 8)
			pixels[offset+1] = byte(g >> 8)
			pixels[offset+2] = byte(r >> 8)
			pixels[offset+3] = byte(a >> 8)

		}
	}

	ximage := C.XCreateImage(
		display,
		C.XDefaultVisual(display, C.XDefaultScreen(display)),
		24,
		C.ZPixmap,
		0,
		(*C.char)(unsafe.Pointer(&pixels[0])),
		C.uint(width),
		C.uint(height),
		32,
		0,
	)

	//defer C.DestroyXImage(ximage) // TODO:

	gc := C.XCreateGC(display, C.Drawable(window), 0, nil)
	defer C.XFreeGC(display, gc) // TODO:

	C.XPutImage(display, C.Drawable(window), gc, ximage, 0, 0, 0, 0, C.uint(width), C.uint(height))
}

func drawBlue(display *C.Display, window C.Window, screen C.int) {
	gc := C.XCreateGC(display, C.Drawable(window), 0, nil)
	defer C.XFreeGC(display, gc)
	colorName := C.CString("blue")
	defer C.free(unsafe.Pointer(colorName))

	var exactColor, screenColor C.XColor
	C.XAllocNamedColor(display, C.XDefaultColormap(display, screen), colorName, &screenColor, &exactColor)

	C.XSetForeground(display, gc, screenColor.pixel)

	C.XFillRectangle(display, C.Drawable(window), gc, 0, 0, width/2, height/2)
}
