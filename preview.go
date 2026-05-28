package golibraw

// #cgo CFLAGS: -I/opt/homebrew/include
// #cgo LDFLAGS: -L/opt/homebrew/lib -lraw
// #include "libraw/libraw.h"
// #include <stdlib.h>
import "C"

import (
	"unsafe"
)

type PreviewFormat int

const (
	PreviewUnknown  PreviewFormat = 0
	PreviewJPEG     PreviewFormat = 1
	PreviewBitmap   PreviewFormat = 2
	PreviewBitmap16 PreviewFormat = 3
	PreviewLayer    PreviewFormat = 4
	PreviewRollei   PreviewFormat = 5
)

type PreviewInfo struct {
	Width  int
	Height int
	Length int // byte length
	Format PreviewFormat
}

func (p *Processor) ExtractLargestPreview(filepath string) ([]byte, PreviewInfo, error) {
	proc, closeProc, err := openRaw(filepath)
	if err != nil {
		return nil, PreviewInfo{}, err
	}
	defer closeProc()

	if err := librawErr(C.libraw_unpack_thumb(proc)); err != nil {
		return nil, PreviewInfo{}, err
	}

	thumb := proc.thumbnail
	if thumb.thumb == nil || thumb.tlength == 0 {
		return nil, PreviewInfo{}, ErrNoPreview
	}

	data := C.GoBytes(unsafe.Pointer(thumb.thumb), C.int(thumb.tlength))
	return data, PreviewInfo{
		Width:  int(thumb.twidth),
		Height: int(thumb.theight),
		Length: int(thumb.tlength),
		Format: PreviewFormat(thumb.tformat),
	}, nil
}

func (p *Processor) PreviewInfoOnly(filepath string) (PreviewInfo, error) {
	proc, closeProc, err := openRaw(filepath)
	if err != nil {
		return PreviewInfo{}, err
	}
	defer closeProc()

	thumb := proc.thumbnail
	if thumb.tlength == 0 && thumb.twidth == 0 && thumb.theight == 0 {
		return PreviewInfo{}, ErrNoPreview
	}
	return PreviewInfo{
		Width:  int(thumb.twidth),
		Height: int(thumb.theight),
		Length: int(thumb.tlength),
		Format: PreviewFormat(thumb.tformat),
	}, nil
}
