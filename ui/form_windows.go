package ui

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procCreateWindowExW  = user32.NewProc("CreateWindowExW")
	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procDefWindowProcW   = user32.NewProc("DefWindowProcW")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procGetMessageW      = user32.NewProc("GetMessageW")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
	procBeginPaint       = user32.NewProc("BeginPaint")
	procEndPaint         = user32.NewProc("EndPaint")
	procFillRect         = user32.NewProc("FillRect")
	procShowWindow       = user32.NewProc("ShowWindow")
	procUpdateWindow     = user32.NewProc("UpdateWindow")
	procLoadCursorW      = user32.NewProc("LoadCursorW")
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	CW_USEDEFAULT       = 0x80000000
	WM_DESTROY          = 0x0002
	WM_PAINT            = 0x000F
	CS_HREDRAW          = 0x0002
	CS_VREDRAW          = 0x0001
	IDC_ARROW           = 32512
	SW_SHOW             = 5
)

type WNDCLASSEX struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}

type MSG struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Point   struct{ X, Y int32 }
}

type PAINTSTRUCT struct {
	HDC         uintptr
	FErase      int32
	RcPaint     struct{ Left, Top, Right, Bottom int32 }
	Restore     int32
	IncUpdate   int32
	RGBReserved [32]byte
}

func Run() {
	hInstance, _, _ := procGetModuleHandleW.Call(0)

	className := syscall.StringToUTF16Ptr("MyWindowClass")
	windowName := syscall.StringToUTF16Ptr("Blue Window")

	wndClass := WNDCLASSEX{
		Size:       uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:      CS_HREDRAW | CS_VREDRAW,
		WndProc:    syscall.NewCallback(windowProc),
		Instance:   hInstance,
		Cursor:     loadCursor(IDC_ARROW),
		Background: 0,
		ClassName:  className,
	}

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wndClass)))

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		WS_OVERLAPPEDWINDOW,
		CW_USEDEFAULT, CW_USEDEFAULT, 800, 600,
		0, 0, hInstance, 0,
	)

	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)

	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func windowProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	/*case WM_PAINT:
	var ps PAINTSTRUCT
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	fillBlue(hdc, &ps.RcPaint)
	procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))*/

	case WM_PAINT:
		var ps PAINTSTRUCT
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

		img, err := loadImageFromEmbed() // Путь к PNG-файлу
		if err == nil {
			drawImageRGBA(hdc, img, &ps.RcPaint)
		}

		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
	default:
		ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return ret
	}
	return 0
}

func fillBlue(hdc uintptr, rect *struct{ Left, Top, Right, Bottom int32 }) {
	blueBrush := createSolidBrush(0xFF0000) // RGB(0, 0, 255) in hex
	defer syscall.Syscall(procFillRect.Addr(), 2, hdc, uintptr(unsafe.Pointer(rect)), uintptr(blueBrush))
}

func createSolidBrush(color uint32) uintptr {
	ret, _, _ := syscall.Syscall(gdi32.NewProc("CreateSolidBrush").Addr(), 1, uintptr(color), 0, 0)
	return ret
}

func loadCursor(cursorID uint16) uintptr {
	cursor, _, _ := procLoadCursorW.Call(0, uintptr(cursorID))
	return cursor
}

func drawImageRGBA(hdc uintptr, img image.Image, rect *struct{ Left, Top, Right, Bottom int32 }) {
	width := rect.Right - rect.Left
	height := rect.Bottom - rect.Top

	compatibleDC, _, _ := gdi32.NewProc("CreateCompatibleDC").Call(hdc)
	defer gdi32.NewProc("DeleteDC").Call(compatibleDC)

	bitmap, _, _ := gdi32.NewProc("CreateCompatibleBitmap").Call(hdc, uintptr(width), uintptr(height))
	defer gdi32.NewProc("DeleteObject").Call(bitmap)

	oldBitmap, _, _ := gdi32.NewProc("SelectObject").Call(compatibleDC, bitmap)
	defer gdi32.NewProc("SelectObject").Call(compatibleDC, oldBitmap)

	bits := make([]byte, width*height*4)
	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			r, g, b, a := img.At(x, y).RGBA()
			offset := (y*int(width) + x) * 4
			bits[offset] = byte(b >> 8)
			bits[offset+1] = byte(g >> 8)
			bits[offset+2] = byte(r >> 8)
			bits[offset+3] = byte(a >> 8)
		}
	}

	bitmapInfo := struct {
		Header struct {
			Size          uint32
			Width         int32
			Height        int32
			Planes        uint16
			BitCount      uint16
			Compression   uint32
			SizeImage     uint32
			XPelsPerMeter int32
			YPelsPerMeter int32
			ClrUsed       uint32
			ClrImportant  uint32
		}
	}{}
	bitmapInfo.Header.Size = uint32(unsafe.Sizeof(bitmapInfo.Header))
	bitmapInfo.Header.Width = int32(width)
	bitmapInfo.Header.Height = -int32(height)
	bitmapInfo.Header.Planes = 1
	bitmapInfo.Header.BitCount = 32
	bitmapInfo.Header.Compression = 0

	gdi32.NewProc("SetDIBits").Call(
		compatibleDC,
		bitmap,
		0,
		uintptr(height),
		uintptr(unsafe.Pointer(&bits[0])),
		uintptr(unsafe.Pointer(&bitmapInfo)),
		0,
	)

	gdi32.NewProc("BitBlt").Call(
		hdc,
		uintptr(rect.Left), uintptr(rect.Top),
		uintptr(width), uintptr(height),
		compatibleDC,
		0, 0, 0x00CC0020, // SRCCOPY
	)
}

//go:embed image.png
var pngContent []byte

func generateImage() (image.Image, error) {
	rgba := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < rgba.Rect.Dy(); y++ {
		for x := 0; x < rgba.Rect.Dx(); x++ {
			rgba.Set(x, y, color.RGBA{0, 0, 255, 255})
		}
	}
	return rgba, nil
}

func loadImageFromEmbed() (image.Image, error) {
	img, err := png.Decode(bytes.NewReader(pngContent))
	return img, err
}
