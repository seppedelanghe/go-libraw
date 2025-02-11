// Package golibraw provides a goroutine‐friendly binding for libraw.
// It wraps key operations (opening, unpacking, processing, and exporting)
// inside a configurable Processor type.
package golibraw

// #cgo CFLAGS: -I/opt/homebrew/include
// #cgo LDFLAGS: -L/opt/homebrew/lib -lraw
// #include "libraw/libraw.h"
// #include <stdlib.h>
import "C"

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"time"
	"unsafe"

	"github.com/lmittmann/ppm"
)

type ImgMetadata struct {
	ScattoTimestamp int64  // capture timestamp
	ScattoDataOra   string // formatted capture date/time
}

type ProcessorOptions struct {}

// Processor is a stateless wrapper for libraw processing.
// Each method creates its own libraw processor so that calls are goroutine‐safe.
type Processor struct {
	options ProcessorOptions
	// TODO: add pool.Sync
}

func NewProcessor(opts ProcessorOptions) *Processor {
	return &Processor{options: opts}
}

func freeCString(s *C.char) {
	C.free(unsafe.Pointer(s))
}

// processFile opens the file, unpacks it, processes it, and returns:
//  - proc: the libraw processor pointer
//  - memImg: the pointer to the in‑memory image returned by libraw_dcraw_make_mem_image
//  - dataSize, height, width, bits: image details
func (p *Processor) processFile(filepath string) (proc *C.libraw_data_t, memImg *C.libraw_processed_image_t, dataSize C.uint,
	height, width, bits C.ushort, err error) {

	proc = C.libraw_init(0)
	if proc == nil {
		err = fmt.Errorf("failed to initialize libraw")
		return
	}

	cFile := C.CString(filepath)
	defer freeCString(cFile)

	ret := C.libraw_open_file(proc, cFile)
	if ret != 0 {
		err = fmt.Errorf("libraw_open_file error: %s", C.GoString(C.libraw_strerror(C.int(ret))))
		C.libraw_close(proc)
		return
	}

	ret = C.libraw_unpack(proc)
	if ret != 0 {
		err = fmt.Errorf("libraw_unpack error: %s", C.GoString(C.libraw_strerror(C.int(ret))))
		C.libraw_close(proc)
		return
	}

	ret = C.libraw_dcraw_process(proc)
	if ret != 0 {
		err = fmt.Errorf("libraw_dcraw_process error: %s", C.GoString(C.libraw_strerror(C.int(ret))))
		C.libraw_close(proc)
		return
	}

	var makeImgErr C.int
	// memImg is a pointer to libraw_processed_image_t.
	memImg = C.libraw_dcraw_make_mem_image(proc, &makeImgErr)
	if makeImgErr != 0 || memImg == nil {
		err = fmt.Errorf("libraw_dcraw_make_mem_image error: %s", C.GoString(C.libraw_strerror(makeImgErr)))
		C.libraw_close(proc)
		return
	}

	dataSize = memImg.data_size
	height = memImg.height
	width = memImg.width
	bits = memImg.bits

	return
}

// clearAndClose releases the memory image and closes the processor.
func clearAndClose(proc *C.libraw_data_t, memImg *C.libraw_processed_image_t) {
	C.libraw_dcraw_clear_mem(memImg)
	C.libraw_recycle(proc)
	C.libraw_close(proc)
}

// ConvertToImage decodes raw image bytes (assumed to be PPM data) into an image.Image.
// It prepends a PPM header to the raw data.
func ConvertToImage(data []byte, width, height, bits int) (image.Image, error) {
	maxVal := (1 << bits) - 1
	header := fmt.Sprintf("P6\n%d %d\n%d\n", width, height, maxVal)
	fullBytes := append([]byte(header), data...)
	return ppm.Decode(bytes.NewReader(fullBytes))
}

// ProcessRaw processes a RAW file and returns an image.Image along with metadata.
func (p *Processor) ProcessRaw(filepath string) (img image.Image, meta ImgMetadata, err error) {
	t0 := time.Now()

	proc, dataPtr, dataSize, height, width, bits, err := p.processFile(filepath)
	if err != nil {
		return nil, ImgMetadata{}, err
	}
	defer clearAndClose(proc, dataPtr)

	dataBytes := C.GoBytes(unsafe.Pointer(&dataPtr), C.int(dataSize))

	img, err = ConvertToImage(dataBytes, int(width), int(height), int(bits))
	if err != nil {
		return nil, ImgMetadata{}, err
	}

	other := C.libraw_get_imgother(proc)
	timestamp := int64(other.timestamp)
	captureTime := time.Unix(timestamp, 0)

	meta = ImgMetadata{
		ScattoTimestamp: timestamp,
		ScattoDataOra:   captureTime.Format("2006-01-02T15:04:05"),
	}
	log.Printf("Processed RAW %s in %v", filepath, time.Since(t0))
	return img, meta, nil
}

