// +build !wasm

package gfx

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func Run(fn func()) {
	pixelgl.Run(fn)
}

func RunTimed(effect TimedEffect) {
	cfg := pixelgl.WindowConfig{
		Title:  "Effect",
		Bounds: pixel.R(0, 0, fRenderWidth*scale, fRenderHeight*scale),
		VSync:  true,
	}
	var (
		frames = 0
		vfps   time.Duration
		update = time.Tick(time.Second / 2)
		lastT  = time.Now()
	)

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	c := win.Bounds().Center()
	started := time.Now()
	bar := pixel.MakePictureData(pixel.R(0, 0, 4, fRenderHeight*scale))
	barRed := pixel.MakePictureData(pixel.R(0, 0, 4, fRenderHeight*scale))
	for i := range bar.Pix {
		bar.Pix[i].G = 255
		bar.Pix[i].A = 192
		barRed.Pix[i].R = 255
		barRed.Pix[i].A = 192
	}
	dst := pixel.MakePictureData(pixel.R(0, 0, fRenderWidth, fRenderHeight))
	var fixedT *float64
	var lastRenderT float64
	for !win.Closed() {
		if win.JustPressed(pixelgl.KeyEscape) {
			win.SetClosed(true)
			continue
		}
		startFrame := time.Now()
		x, _ := QueryPerformanceCounter()
		t := float64(startFrame.Sub(started)) / float64(time.Second*10)
		fixedT = updateInput(win, fixedT, lastRenderT)
		if fixedT != nil {
			t = *fixedT
		}
		_, t = math.Modf(t)
		if t < 0 {
			t = 1 - t
		}
		lastRenderT = t
		pic := effect.Render(t)
		spent := time.Now().Sub(startFrame)
		y, err := QueryPerformanceCounter()
		if err == nil {
			f, err := QueryPerformanceFreq()
			if err == nil {
				spent = (time.Duration(y-x) * time.Second) / time.Duration(f)
			}
		}
		vfps += spent
		elapsed := float64(spent) / float64(time.Second/vSync)
		elapsed = math.Min(elapsed, 1)

		copyTo(dst, pic)
		pixel.NewSprite(dst, dst.Bounds()).
			Draw(win, pixel.IM.Moved(c).Scaled(c, scale))

		// Draw vsync bar
		if elapsed < 1 {
			tl := win.Bounds().H() - (fRenderHeight*elapsed*scale)/2
			pixel.NewSprite(bar, pixel.R(0, 0, 4, scale*fRenderHeight*elapsed)).
				Draw(win, pixel.IM.Moved(pixel.Vec{2, tl}))
		} else {
			tl := win.Bounds().H() - (fRenderHeight*elapsed*scale)/2
			pixel.NewSprite(barRed, pixel.R(0, 0, 4, fRenderHeight*scale*elapsed)).
				Draw(win, pixel.IM.Moved(pixel.Vec{2, tl}))
		}
		win.Update()
		frames++
		select {
		case <-update:
			elapsed := time.Since(lastT)
			virtual := time.Duration(frames) * elapsed / vfps
			virtual = virtual * time.Second / elapsed
			rframes := float64(frames) * float64(time.Second) / float64(elapsed)
			win.SetTitle(fmt.Sprintf("%s | time: %0.3f | FPS: %.0f | vFPS: %d", cfg.Title, lastRenderT, rframes, virtual))
			frames = 0
			vfps = 0
			lastT = time.Now()
		default:
		}
	}
}

func RunWriteToDisk(fx TimedEffect, n int, path string) {
	const length = 10 * vSync
	frame := 0
	type toSave struct {
		img image.Image
		fn  string
	}
	dir, err := filepath.Abs(path)
	if err == nil {
		os.MkdirAll(filepath.Dir(dir), os.ModePerm)
	}
	save := make(chan toSave)
	var wg sync.WaitGroup
	wg.Add(8)
	for i := 0; i < 8; i++ {
		go func() {
			defer wg.Done()
			for img := range save {
				func() {
					f, err := os.Create(img.fn)
					if err != nil {
						panic(err)
					}
					defer f.Close()
					err = png.Encode(f, img.img)
					if err != nil {
						panic(err)
					}
				}()
			}
		}()
	}
	for i := 0; i < n; i++ {
		for j := 0; j < length; j++ {
			img := fx.Render(float64(j) / length)
			switch i := img.(type) {
			case *image.Gray:
				save <- toSave{
					img: copyGrayToPaletted(i),
					fn:  fmt.Sprintf(path, frame),
				}
			case *image.Paletted:
				dst := image.NewPaletted(i.Rect, i.Palette)
				for y := 0; y < dst.Rect.Dy(); y++ {
					copy(dst.Pix[y*dst.Stride:y*dst.Stride+dst.Rect.Dx()], i.Pix[y*i.Stride:y*i.Stride+i.Rect.Dx()])
				}
				save <- toSave{
					img: dst,
					fn:  fmt.Sprintf(path, frame),
				}
			default:
				dst := image.NewRGBA(img.Bounds())
				draw.Draw(dst, i.Bounds(), img, image.Pt(0, 0), draw.Over)
				save <- toSave{
					img: dst,
					fn:  fmt.Sprintf(path, frame),
				}
			}
			frame++
			fmt.Printf(path+"\n", frame)
		}
	}
	close(save)
	wg.Wait()
}

func updateInput(win *pixelgl.Window, t *float64, lastT float64) *float64 {
	var fP = func(f float64) *float64 { return &f }
	switch {
	case win.JustPressed(pixelgl.Key0):
		return fP(0)
	case win.JustPressed(pixelgl.Key1):
		return fP(1.0 / 10)
	case win.JustPressed(pixelgl.Key2):
		return fP(2.0 / 10)
	case win.JustPressed(pixelgl.Key3):
		return fP(3.0 / 10)
	case win.JustPressed(pixelgl.Key4):
		return fP(4.0 / 10)
	case win.JustPressed(pixelgl.Key5):
		return fP(5.0 / 10)
	case win.JustPressed(pixelgl.Key6):
		return fP(6.0 / 10)
	case win.JustPressed(pixelgl.Key7):
		return fP(7.0 / 10)
	case win.JustPressed(pixelgl.Key8):
		return fP(8.0 / 10)
	case win.JustPressed(pixelgl.Key9):
		return fP(9.0 / 10)
	case win.JustPressed(pixelgl.KeyA):
		return fP(999.9 / 1000)
	case win.Pressed(pixelgl.KeyLeft):
		d := 1.0 / 1000
		if win.Pressed(pixelgl.KeyLeftShift) || win.Pressed(pixelgl.KeyRightShift) {
			d *= 0.1
		}
		return fP(lastT - d)
	case win.Pressed(pixelgl.KeyRight):
		d := 1.0 / 1000
		if win.Pressed(pixelgl.KeyLeftShift) || win.Pressed(pixelgl.KeyRightShift) {
			d *= 0.1
		}
		return fP(lastT + d)
	case win.JustPressed(pixelgl.KeySpace):
		if t == nil {
			return &lastT
		}
		return nil
	}
	return t
}
func copyTo(dst *pixel.PictureData, src image.Image) {
	switch s := src.(type) {
	case *image.Paletted:
		copyToPaletted(dst, s)
	case *image.Gray:
		copyToGray(dst, s)
	default:
		copyToGen(dst, src)
	}
}

func copyToPaletted(dst *pixel.PictureData, src *image.Paletted) {
	var pal [256]color.RGBA
	for i, col := range src.Palette {
		r, g, b, a := col.RGBA()
		pal[i] = color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: uint8(a >> 8),
		}
	}

	for y := 0; y < src.Rect.Dy(); y++ {
		line := src.Pix[y*src.Stride : y*src.Stride+src.Rect.Dx()]
		// Destination is flipped.
		dstY := src.Rect.Dy() - y - 1
		dLine := dst.Pix[dstY*dst.Stride : dstY*dst.Stride+len(line)]
		for x, v := range line {
			dLine[x] = pal[v]
		}
	}
}

func copyToGray(dst *pixel.PictureData, src *image.Gray) {
	for y := 0; y < src.Rect.Dy(); y++ {
		line := src.Pix[y*src.Stride : y*src.Stride+src.Rect.Dx()]
		// Destination is flipped.
		dstY := src.Rect.Dy() - y - 1
		dLine := dst.Pix[dstY*dst.Stride : dstY*dst.Stride+len(line)]
		for x, v := range line {
			dLine[x] = paletteRGBA[v]
		}
	}
}

func copyToRGBAGray(src *image.Gray) *image.RGBA {
	dst := image.NewRGBA(src.Rect)
	for y := 0; y < src.Rect.Dy(); y++ {
		line := src.Pix[y*src.Stride : y*src.Stride+src.Rect.Dx()]
		dLine := dst.Pix[y*dst.Stride : y*dst.Stride+dst.Stride]
		for x, v := range line {
			p := palette[v]
			dLine[x*4+0] = byte(p)
			dLine[x*4+1] = byte(p >> 8)
			dLine[x*4+2] = byte(p >> 16)
			dLine[x*4+3] = 0xff
		}
	}
	return dst
}

func copyGrayToPaletted(src *image.Gray) *image.Paletted {
	pal := make(color.Palette, 256)
	for i := range pal {
		v := palette[i]
		pal[i] = color.RGBA{R: byte(v), G: byte(v >> 8), B: byte(v >> 16), A: 255}
	}
	dst := image.NewPaletted(src.Rect, pal)
	for y := 0; y < src.Rect.Dy(); y++ {
		line := src.Pix[y*src.Stride : y*src.Stride+src.Rect.Dx()]
		dLine := dst.Pix[y*dst.Stride : y*dst.Stride+dst.Stride]
		copy(dLine, line)
	}
	return dst
}

// copyToGen, allocates... Implement faster ones...
func copyToGen(dst *pixel.PictureData, src image.Image) {
	rgba := image.NewRGBA(src.Bounds())
	draw.Draw(rgba, rgba.Bounds(), src, image.Pt(0, 0), draw.Src)

	for y := 0; y < rgba.Rect.Dy(); y++ {
		line := rgba.Pix[y*rgba.Stride : y*rgba.Stride+rgba.Rect.Dx()]
		// Destination is flipped.
		dstY := rgba.Rect.Dy() - y - 1
		dLine := dst.Pix[dstY*dst.Stride : dstY*dst.Stride+len(line)]
		for x := range dLine {
			dLine[x] =
				color.RGBA{
					R: uint8(line[x*4] >> 8),
					G: uint8(line[x*4+1] >> 8),
					B: uint8(line[x*4+2] >> 8),
					A: uint8(line[x*4+3] >> 8),
				}
		}
	}
}
