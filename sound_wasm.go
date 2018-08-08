// +build wasm

package gfx

import (
	"fmt"
	"time"
)

type soundPlayer struct {
	s         jsObject
	startedAt time.Time
}

func loadMusic(path string) (MusicPlayer, error) {
	m := soundPlayer{s: getElementById("sound")}
	return &m, nil
}

func (m *soundPlayer) Start(cb func(duration time.Duration)) {
	m.startedAt = time.Now()
	res := m.s.Call("play")
	fmt.Printf("%+v, %#v\n", res, res)
	cb(0)
}

func (m *soundPlayer) Pos() time.Duration {
	return time.Since(m.startedAt)
}
