package gfx

import "time"

type MusicPlayer interface {
	Start(func(duration time.Duration))
	Pos() time.Duration
}
