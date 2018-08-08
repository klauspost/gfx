package gfx

import (
	"fmt"
	"os"
)

type LoadFn func(name string) ([]byte, error)

var loaders []LoadFn

// AddData will add a data loader.
func AddData(fn LoadFn) {
	loaders = append(loaders, fn)
}

func Load(file string) ([]byte, error) {
	for i, fn := range loaders {
		b, err := fn(file)
		if err == nil {
			return b, nil
		}
		fmt.Println("Loader", i, "returned", err.Error())
	}
	return nil, os.ErrNotExist
}
