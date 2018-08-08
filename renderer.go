package gfx

import "image"

type Options struct {
	ScreenSize image.Rectangle
}

type TimedEffect interface {
	Render(t float64) image.Image
}

type ProgressiveEffect interface {
	Reset(o Options)
	Render() image.Image
}