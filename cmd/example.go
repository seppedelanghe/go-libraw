package main

import (
	"image/jpeg"
	"os"

	libraw "github.com/seppedelanghe/go-libraw"
)


const RawPath = "testdata/sample.NEF"

func main() {
	processor := libraw.NewProcessor(libraw.ProcessorOptions{})
	img, _, err := processor.ProcessRaw(RawPath)
	if err != nil {
		panic(err)
	}

	file, err := os.Create("output.jpg")
	if err != nil {
		panic(err)
	}
	
	err = jpeg.Encode(file, img, &jpeg.Options{
		Quality: jpeg.DefaultQuality,
	})
	if err != nil {
		panic(err)
	}
}
