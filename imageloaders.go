package gfx

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
)

func LoadPalPicture(path string) (*image.Paletted, error) {
	dat, err := Load(path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewBuffer(dat))
	if err != nil {
		return nil, err
	}
	ipi, ok := img.(*image.Paletted)
	if !ok {
		gray, ok := img.(*image.Gray)
		if !ok {
			return nil, fmt.Errorf("unexpected type %T", img)
		}
		pal := make(color.Palette, 256)
		for i := range pal {
			pal[i] = color.Gray{uint8(i)}
		}
		ipi = image.NewPaletted(gray.Rect, pal)
		draw.Draw(ipi, ipi.Rect, gray, image.Pt(0, 0), draw.Src)
	}
	return ipi, nil
}

func LoadGreyPicture(path string) (*image.Gray, error) {
	dat, err := Load(path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewBuffer(dat))
	if err != nil {
		return nil, err
	}
	gray, ok := img.(*image.Gray)
	if ok {
		return gray, nil
	}
	return ToGray(img), nil
}
