//+build !windows

package gfx

import "errors"

func QueryPerformanceCounter() (int64, error) {
	return 0, errors.New("unavailable")
}

func QueryPerformanceFreq() (int64, error) {
	return 0, errors.New("unavailable")
}
