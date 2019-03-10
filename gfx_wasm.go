// +build wasm

package gfx

import (
	"fmt"
	"image"
	"log"
	"math"
	"runtime/debug"
	"syscall/js"
	"time"
)

type jsObject = js.Value

var (
	Global      = js.Global()
	newCallback = js.NewCallback
	status      = js.Undefined()
	document    jsObject
)

func setStatus(s string) {
	if status != js.Undefined() {
		status.Set("innerHTML", s)
	} else {
		// Print to console
		fmt.Println(s)
	}
}

func getElementById(name string) jsObject {
	node := document.Call("getElementById", name)
	if isUndefined(node) {
		log.Fatalf("Couldn't find element %q", name)
	}
	return node
}

func isUndefined(v js.Value) bool {
	return v == js.Undefined()
}

func Run(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			setStatus(fmt.Sprintf("ERROR: %v", r))
			debug.PrintStack()
		}
	}()
	document = Global.Get("document")
	if isUndefined(document) {
		log.Fatalf("Didn't find document - not running in browser")
	}
	Wait()
	fn()
	select {}
}

func Wait() {
	pic, err := LoadGreyPicture("data/click.png")
	if err != nil {
		panic(err)
	}
	status = getElementById("status")
	setStatus("Initializing...")
	canvas := getElementById("fx-display")
	ctx := canvas.Call("getContext", "2d")
	canvasData := ctx.Call("createImageData", renderWidth, renderHeight)
	data := canvasData.Get("data")
	screen32 := make([]byte, renderWidth*renderHeight*4)
	copyToGrey(screen32, pic)
	data.Call("set", js.TypedArrayOf(screen32))
	ctx.Call("putImageData", canvasData, 0, 0)

	var cb js.Callback
	ready := make(chan struct{})
	cb = js.NewCallback(func(args []js.Value) {
		fmt.Println("button clicked")
		close(ready)
		cb.Release() // release the callback if the button will not be clicked again
		canvas.Call("removeEventListener", "click", cb)
	})
	canvas.Call("addEventListener", "click", cb)
	setStatus("Waiting for start")
	<-ready
}

func RunTimedDur(fx TimedEffect, duration time.Duration) {
	canvas := getElementById("fx-display")
	ctx := canvas.Call("getContext", "2d")
	canvasData := ctx.Call("createImageData", renderWidth, renderHeight)
	data := canvasData.Get("data")

	screen32 := make([]byte, renderWidth*renderHeight*4)
	const printInterval = vSync
	var (
		fixedT      *float64
		lastRenderT float64
		vfps        time.Duration
		lastT       = time.Now()
		started     = time.Now()
		frames      int
	)

	var draw func(args []jsObject)
	draw = func(args []jsObject) {
		defer func() {
			if r := recover(); r != nil {
				setStatus(fmt.Sprintf("ERROR: %v", r))
				debug.PrintStack()
			}
		}()
		startFrame := time.Now()
		t := float64(startFrame.Sub(started)) / float64(duration)
		fixedT = updateInput(fixedT, lastRenderT)
		if fixedT != nil {
			t = *fixedT
		}
		_, t = math.Modf(t)
		if t < 0 {
			t = 1 - t
		}
		lastRenderT = t
		screen := fx.Render(t)
		spent := time.Now().Sub(startFrame)
		vfps += spent
		elapsed := float64(spent) / float64(time.Second/vSync)
		elapsed = math.Min(elapsed, 1)
		switch scr := screen.(type) {
		case *image.Paletted:
			copyToPaletted(screen32, scr)
		case *image.Gray:
			copyToGrey(screen32, scr)
		case *image.RGBA:
			copyToRGBA(screen32, scr)
		default:
			panic(fmt.Sprintf("Unhandled screen type: %T", screen))
		}
		if elapsed < 1 {
			h := int(elapsed * float64(renderHeight))
			for i := 0; i < h; i++ {
				p := i * renderWidth * 4
				screen32[p] = 0
				screen32[p+1] = 0xff
				screen32[p+2] = 0
				screen32[p+4] = 0
				screen32[p+5] = 0xff
				screen32[p+6] = 0
			}
		} else {
			for i := 0; i < renderHeight; i++ {
				p := i * renderWidth * 4
				screen32[p] = 0xff
				screen32[p+1] = 0
				screen32[p+2] = 0
				screen32[p+4] = 0xff
				screen32[p+5] = 0
				screen32[p+6] = 0
			}
		}

		data.Call("set", js.TypedArrayOf(screen32))
		ctx.Call("putImageData", canvasData, 0, 0)

		frames++
		if frames >= printInterval {
			elapsed := time.Since(lastT)
			virtual := time.Duration(frames) * elapsed / vfps
			virtual = virtual * time.Second / elapsed
			rframes := float64(frames) * float64(time.Second) / float64(elapsed)
			setStatus(fmt.Sprintf("%s | time: %0.3f | FPS: %.0f | vFPS: %d", "FX", lastRenderT, rframes, virtual))

			frames = 0
			vfps = 0
			lastT = time.Now()
		}
		Global.Call("requestAnimationFrame", newCallback(draw))
	}
	Global.Call("requestAnimationFrame", newCallback(draw))
}

func copyToPaletted(dst []byte, src *image.Paletted) {
	var pal [256][4]byte
	for i, col := range src.Palette {
		r, g, b, _ := col.RGBA()
		pal[i] = [4]byte{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 0xff}
	}
	w, h := src.Rect.Dx(), src.Rect.Dy()

	for y := 0; y < h; y++ {
		line := src.Pix[y*src.Stride : y*src.Stride+w]
		dstY := y * w * 4
		dLine := dst[dstY : dstY+len(line)*4]
		for x, v := range line {
			w := pal[v]
			dLine[x*4+3] = 0xff
			dLine[x*4+2] = w[2]
			dLine[x*4+1] = w[1]
			dLine[x*4+0] = w[0]
		}
	}
}

func copyToGrey(dst []byte, src *image.Gray) {
	w, h := src.Rect.Dx(), src.Rect.Dy()
	for y := 0; y < h; y++ {
		line := src.Pix[y*src.Stride : y*src.Stride+w]
		dstY := y * w * 4
		dLine := dst[dstY : dstY+len(line)*4]
		for x, v := range line {
			p := palette[v]
			dLine[x*4+0] = byte(p)
			dLine[x*4+1] = byte(p >> 8)
			dLine[x*4+2] = byte(p >> 16)
			dLine[x*4+3] = 0xff
		}
	}
}

func copyToRGBA(dst []byte, src *image.RGBA) {
	w, h := src.Rect.Dx(), src.Rect.Dy()
	for y := 0; y < h; y++ {
		line := src.Pix[y*src.Stride : y*src.Stride+w*4]
		dstY := y * w * 4
		dLine := dst[dstY : dstY+len(line)]
		copy(dLine, line)
	}
}

func updateInput(t *float64, lastT float64) *float64 {
	return t
}
