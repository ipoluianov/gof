package ui

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/Xatom.h>
#include <stdlib.h>

void DestroyXImage(XImage *image) {
	image->f.destroy_image(image);
}

*/
import "C"
import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"
	"time"
	"unsafe"
)

var posX C.uint
var posY C.uint
var width C.uint
var height C.uint

func resizeWindow(display *C.Display, window C.Window, width, height int) {
	C.XResizeWindow(display, window, C.uint(width), C.uint(height))
	fmt.Printf("Window resized to %dx%d\n", width, height)
}

func moveWindow(display *C.Display, window C.Window, x, y int) {
	C.XMoveWindow(display, window, C.int(x), C.int(y))
	fmt.Printf("Window moved to (%d, %d)\n", x, y)
}

func setWindowTitle(display *C.Display, window C.Window, title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.XStoreName(display, window, cTitle)
	fmt.Printf("Window title set to: %s\n", title)
}

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

	C.XSelectInput(display, window, C.ExposureMask|C.PropertyChangeMask|C.ResizeRedirectMask|C.StructureNotifyMask|C.KeyPressMask|C.StructureNotifyMask|C.KeyReleaseMask|C.EnterWindowMask|C.LeaveWindowMask|C.ButtonPressMask|C.ButtonReleaseMask|C.PointerMotionMask)

	C.XMapWindow(display, window)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		for C.XPending(display) > 0 {
			var event C.XEvent
			C.XNextEvent(display, &event)

			switch eventType(event) {

			case C.Expose:
				drawBlue(display, window, screen)
				img, err := loadImageFromEmbed()
				if err == nil {
					drawImageRGBA(display, window, img)
				}
			case C.MapNotify:
				mapEvent := (*C.XMapEvent)(unsafe.Pointer(&event))
				fmt.Printf("Window became visible. Window ID: %d\n", mapEvent.window)

			case C.UnmapNotify:
				unmapEvent := (*C.XUnmapEvent)(unsafe.Pointer(&event))
				fmt.Printf("Window was hidden. Window ID: %d\n", unmapEvent.window)
			case C.DestroyNotify:
				destroyEvent := (*C.XDestroyWindowEvent)(unsafe.Pointer(&event))
				fmt.Printf("Window was destroyed. Window ID: %d\n", destroyEvent.window)

			case C.ReparentNotify:
				reparentEvent := (*C.XReparentEvent)(unsafe.Pointer(&event))
				fmt.Printf("Window changed parent. Window ID: %d, New Parent ID: %d\n", reparentEvent.window, reparentEvent.parent)
			case C.ResizeRequest:
				resizeEvent := (*C.XResizeRequestEvent)(unsafe.Pointer(&event))
				fmt.Printf("Resize request received: Width=%d, Height=%d\n", resizeEvent.width, resizeEvent.height)

			case C.ConfigureNotify:
				configureEvent := (*C.XConfigureEvent)(unsafe.Pointer(&event))
				posX = C.uint(configureEvent.x)
				posY = C.uint(configureEvent.y)
				width = C.uint(configureEvent.width)
				height = C.uint(configureEvent.height)
				fmt.Println("Configure:", posX, posY, width, height)

			case C.KeyPress:
				keyEvent := (*C.XKeyEvent)(unsafe.Pointer(&event))
				keySym := C.XLookupKeysym((*C.XKeyEvent)(unsafe.Pointer(&event)), 0)
				fmt.Printf("Key pressed: KeySym = %d, KeyCode = %d\n", keySym, keyEvent.keycode)

				//resizeWindow(display, window, 600, 200)
				//moveWindow(display, window, 100, 100)
				//setWindowTitle(display, window, "HELLO")
			case C.KeyRelease:
				keyEvent := (*C.XKeyEvent)(unsafe.Pointer(&event))
				keySym := C.XLookupKeysym(keyEvent, 0)
				fmt.Printf("Key released: KeySym = %d, KeyCode = %d\n", keySym, keyEvent.keycode)
			case C.EnterNotify:
				enterEvent := (*C.XCrossingEvent)(unsafe.Pointer(&event))
				fmt.Printf("Cursor entered window at (%d, %d)\n", enterEvent.x, enterEvent.y)

			case C.LeaveNotify:
				leaveEvent := (*C.XCrossingEvent)(unsafe.Pointer(&event))
				fmt.Printf("Cursor left window at (%d, %d)\n", leaveEvent.x, leaveEvent.y)
			case C.MotionNotify:
				motionEvent := (*C.XMotionEvent)(unsafe.Pointer(&event))
				fmt.Printf("Mouse moved to (%d, %d)\n", motionEvent.x, motionEvent.y)

			case C.ButtonPress:
				buttonEvent := (*C.XButtonEvent)(unsafe.Pointer(&event))
				fmt.Printf("Mouse button %d pressed at (%d, %d)\n", buttonEvent.button, buttonEvent.x, buttonEvent.y)

			case C.ButtonRelease:
				buttonEvent := (*C.XButtonEvent)(unsafe.Pointer(&event))
				fmt.Printf("Mouse button %d released at (%d, %d)\n", buttonEvent.button, buttonEvent.x, buttonEvent.y)
			}
		}

		select {
		case <-ticker.C:
			//fmt.Println("Timer event: 10ms tick")
		default:
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
