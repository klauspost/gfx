// +build !wasm

package gfx

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type mp3Player struct {
	streamer  beep.StreamSeekCloser
	format    beep.Format
	file      *os.File
	startedAt time.Time
}

func loadMusic(path string) (MusicPlayer, error) {
	fp, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	m := mp3Player{}
	m.file, err = os.Open(fp)
	if err != nil {
		return nil, err
	}
	m.streamer, m.format, err = mp3.Decode(m.file)
	if err != nil {
		return nil, err
	}
	err = speaker.Init(m.format.SampleRate, m.format.SampleRate.N(time.Second/10))
	if err != nil {
		return nil, err
	}
	fmt.Printf("Playing %#v\n", m.format)
	return &m, nil
}

func (m *mp3Player) Start(cb func(duration time.Duration)) {
	speaker.Play(beep.Seq(m.streamer, beep.Callback(func() {
		// Callback after the stream Ends
		fmt.Println("done")
	})))
	// We have no reasonable way to get info on exactly when playback started.
	// We can only see side effects on the stream, so we assume it started *now*
	m.startedAt = time.Now()
	if false {
		go func() {
			t := time.NewTicker(time.Second)
			for {
				select {
				case <-t.C:
					m.Pos()
				}
			}
		}()
	}
	cb(0)
}

func (m *mp3Player) Pos() time.Duration {
	if false {
		p := m.streamer.Position()
		d := time.Second * time.Duration(p) / time.Duration(m.format.SampleRate)
		fmt.Println("pos:", p, "dur:", d, "since start:", time.Since(m.startedAt))
	}
	return time.Since(m.startedAt)
}
