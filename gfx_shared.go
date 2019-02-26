package gfx

import (
	"image"
	"image/color"
	"image/draw"
	"time"
)

var (
	renderWidth   = 640
	renderHeight  = 360
	fRenderWidth  = 640.0
	fRenderHeight = 360.0
	fullscreen    = false
)

func SetRenderSize(w, h int) {
	renderWidth, renderHeight = w, h
	fRenderWidth, fRenderHeight = float64(renderWidth), float64(renderHeight)
}

func Fullscreen(b bool) {
	fullscreen = b
}

const (
	scale = 2.0
	vSync = 60
)

func RunTimed(effect TimedEffect) {
	RunTimedDur(effect, 10*time.Second)
}

func RunTimedMusic(effect TimedEffect, musicFile string) {
	sfx, err := loadMusic(musicFile)
	if err != nil {
		panic(err)
	}
	sfx.Start(func(duration time.Duration) {
		RunTimed(effect)
	})
}

func ToGray(img image.Image) *image.Gray {
	grey := image.NewGray(img.Bounds())
	draw.Draw(grey, grey.Rect, img, image.Pt(0, 0), draw.Src)
	return grey
}

var palette [256]uint32
var paletteRGBA [256]color.RGBA

func init() {
	for i := range palette[:] {
		c := uint32(i)
		c |= c << 8
		c |= c << 16
		c |= 255 << 24
		palette[i] = c
		paletteRGBA[i] = color.RGBA{byte(c), byte(c >> 8), byte(c >> 16), 255}
	}
}

func InitGreyPalette(p [256]uint32) {
	for i, c := range p[:] {
		palette[i] = c
		paletteRGBA[i] = color.RGBA{byte(c), byte(c >> 8), byte(c >> 16), 255}
	}
}

// InitGreyShadedPalette will initialize a palette that goes from
// black -> neutral -> white.
// Specify the palette index where the color should be the supplied value.
func InitShadedPalette(neutral int, rgb color.RGBA) {
	var palette [256]uint32
	rC, gC, bC := int(rgb.R), int(rgb.G), int(rgb.B)
	flipScale := int(256.0 * (256.0 / float64(neutral)))
	flipScale2 := int(256.0 * (256.0 / float64(255-neutral)))
	for i := range palette[:] {
		var r, g, b int
		if i < neutral {
			r = (i*rC*flipScale + 128) >> 16
			g = (i*gC*flipScale + 128) >> 16
			b = (i*bC*flipScale + 128) >> 16
		} else {
			r = rC
			g = gC
			b = bC
			r += ((i - neutral) * (255 - rC) * flipScale2) >> 16
			g += ((i - neutral) * (255 - gC) * flipScale2) >> 16
			b += ((i - neutral) * (255 - bC) * flipScale2) >> 16
		}
		palette[i] = uint32(r) | (uint32(g) << 8) | (uint32(b) << 16)
	}
	InitGreyPalette(palette)
}
