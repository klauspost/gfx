package gfx

import (
	"image"
	"image/color"
	"image/draw"
	"time"
)

const (
	renderWidth  = 640
	renderHeight = 360
	scale        = 2.0
	vSync        = 60
)

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
