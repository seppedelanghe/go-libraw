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
	"fmt"
	"image"
	"log"
	"time"
	"unsafe"
)

type ImgMetadata struct {
	CaptureTimestamp int64
	CaptureDate time.Time
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


func ConvertToImage(data []byte, width, height, bits int) (image.Image, error) {
    // Check if we have the expected amount of data for RGB
    expectedSize := width * height * 3 // 3 bytes per pixel for RGB
    if len(data) != expectedSize {
        return nil, fmt.Errorf("unexpected data size: got %d, want %d", len(data), expectedSize)
    }

    // Create a new RGB image
    img := image.NewRGBA(image.Rect(0, 0, width, height))
    
    // Convert the raw RGB data to RGBA
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            offset := (y*width + x) * 3 // 3 bytes per pixel in source
            r := data[offset]
            g := data[offset+1]
            b := data[offset+2]
            
            // Set pixel in the RGBA image
            dstOffset := (y*width + x) * 4 // 4 bytes per pixel in RGBA
            img.Pix[dstOffset] = r
            img.Pix[dstOffset+1] = g
            img.Pix[dstOffset+2] = b
            img.Pix[dstOffset+3] = 255 // Alpha channel
        }
    }
    
    return img, nil
}

// ProcessRaw processes a RAW file and returns an image.Image along with metadata.
func (p *Processor) ProcessRaw(filepath string) (img image.Image, meta ImgMetadata, err error) {
    t0 := time.Now()

    proc, dataPtr, dataSize, height, width, bits, err := p.processFile(filepath)
    if err != nil {
        return nil, ImgMetadata{}, err
    }
    defer clearAndClose(proc, dataPtr)

    // Convert raw bytes to Go slice
    dataBytes := C.GoBytes(unsafe.Pointer(&dataPtr.data[0]), C.int(dataSize))

    // Handle different bit depths
    if bits > 8 {
        // Convert higher bit depth to 8-bit
        adjustedData := make([]byte, width*height*3)
        for i := 0; i < len(dataBytes); i += 2 {
            // Combine two bytes into one, shifting to 8-bit depth
            if i+1 < len(dataBytes) {
                value := uint16(dataBytes[i]) | (uint16(dataBytes[i+1]) << 8)
                adjustedData[i/2] = byte(value >> (bits - 8))
            }
        }
        dataBytes = adjustedData
    }

    img, err = ConvertToImage(dataBytes, int(width), int(height), 8)
    if err != nil {
        return nil, ImgMetadata{}, fmt.Errorf("convert to image: %v", err)
    }

    other := C.libraw_get_imgother(proc)
    timestamp := int64(other.timestamp)
    captureTime := time.Unix(timestamp, 0)

    meta = ImgMetadata{
		CaptureTimestamp: timestamp,
        CaptureDate: captureTime,
    }
    log.Printf("Processed RAW %s in %v", filepath, time.Since(t0))
    return img, meta, nil
}

