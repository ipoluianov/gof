package main

import (
	"runtime"

	"github.com/ipoluianov/gof/ui"
)

func main() {
	runtime.LockOSThread()
	ui.Run()
}
