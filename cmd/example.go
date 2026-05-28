package main

import (
	"errors"
	"fmt"
	"image/jpeg"
	"os"

	libraw "github.com/seppedelanghe/go-libraw"
)

const RawPath = "testdata/_SPC2147.NEF"

func main() {
	processor := libraw.NewProcessor(libraw.NewProcessorOptions())
	img, metadata, err := processor.ProcessRaw(RawPath)
	if err != nil {
		panic(err)
	}

	_, info, err := processor.ExtractLargestPreview(RawPath)
	if err != nil {
		if errors.Is(err, libraw.ErrNoPreview) {
			fmt.Println("No preview image present")
		} else {
			panic(err)
		}
	} else {
		fmt.Printf("Image has embedded preview. Size (h x w): %d x %d\n", info.Height, info.Width)
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

	fmt.Printf("Camera make: %s\nImage size (h x w): %d x %d\n", metadata.IData.Make, metadata.Sizes.Height, metadata.Sizes.Width)
}
