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

type OutputColor uint8

const (
	Raw OutputColor = iota
	SRGB
	AdobeRGB
	WideGamutRGB
	ProPhotoRGB
	XYZ
	ACES
	DciP3
	Rec2020
)


type Box struct {
	X1 uint // x1 or x
	Y1 uint // y1 or y

	X2 uint // x2 or w
	Y2 uint // y2 or h
}

func (b *Box) toC() [4]C.uint {
    return [4]C.uint{C.uint(b.X1), C.uint(b.Y1), C.uint(b.X2), C.uint(b.Y2)}
}

func (b *Box) IsEmpty() bool {
	return b.X1 == 0 && b.Y1 == 0 && b.X2 == 0 && b.Y2 == 0
}

type ProcessorOptions struct {
	Greybox Box // coordinates (in pixels) of the rectangle that is used to calculate the white balance
	Cropbox Box // image cropping re ctangle
	Aber [4]float64 // correction of chromatic aberrations
	Gamm [6]float64 // user gamma-curve
	UserMul [4]float32 // 4 multipliers (r,g,b,g) of the user's white balance
	Bright float32
	Threshold float32 // threshold for wavelet denoising

	HalfSize bool // output image at 50% size
	FourColorRGB bool // switches on separate interpolations for two green components
	Highlight int // 0-9: Highlight mode (0=clip, 1=unclip, 2=blend, 3+=rebuild)
	UseAutoWb bool
	UseCameraWb bool
	UseCameraMatrix int

	OutputColor OutputColor
	
	OutputProfile string // path to output profile ICC file
	CameraProfile string // path to input profile ICC file, or 'embed' for embedded profile
	BadPixels string // path to bad pixels map file
	DarkFrame string // path to dark frame file
	
	OutputBps int // 8 or 16
	OutputTiff bool
	OutputFlags int // Bitfield that allows to set output file options
	UserFlip int // EXIF rotation flags -> 0 = no rotation
	UserQual int // Interpolaton -> 0 = Linear, 1 = VNG, 2 = PPG, 3 = AHD
	UserBlack int
	UserCblack [4]int // per-channel black level offsets
	UserSat int // Saturation
	MedPasses int // median filter passes for noise reduction
	AutoBrightThr float32 // Threshold for auto-brightness correction
	AdjustMaximumThr float32 // Threshold for adjusting maximum brigtness value in auto-exposure calculations
	NoAutoBright bool // 0 = enable auto-brightness, 1 = disabled
	UseFujiRotate bool // 1 apply Fuji sensor rotation
	GreenMatching bool // Enable green channel equalization
	DcbIterations int
	DcbEnhanceFl bool
	FbddNoiserd int // 0 = do not use, 1 = light reduction, 2 full reduction
	ExpCorrect bool
	ExpShift float32
	ExpPreser float32
	NoAutoScale bool
	NoInterpolation bool
}

func (opts *ProcessorOptions) bool(v bool) C.int {
	if v {
		return C.int(1)
	}
	return C.int(0)
}

func (opts *ProcessorOptions) Apply(params C.libraw_output_params_t) C.libraw_output_params_t {
	if !opts.Greybox.IsEmpty() {
		params.greybox = opts.Greybox.toC()
	}
	if !opts.Cropbox.IsEmpty() {
		params.cropbox = opts.Cropbox.toC()
	}

	params.aber = [4]C.double{
		C.double(opts.Aber[0]),
		C.double(opts.Aber[1]),
		C.double(opts.Aber[2]),
		C.double(opts.Aber[3]),
	}
	params.gamm = [6]C.double{
		C.double(opts.Gamm[0]),
		C.double(opts.Gamm[1]),
		C.double(opts.Gamm[2]),
		C.double(opts.Gamm[3]),
		C.double(opts.Gamm[4]),
		C.double(opts.Gamm[5]),
	}
	params.user_mul = [4]C.float{
		C.float(opts.UserMul[0]),
		C.float(opts.UserMul[1]),
		C.float(opts.UserMul[2]),
		C.float(opts.UserMul[3]),
	}
	params.bright = C.float(opts.Bright)
	params.threshold = C.float(opts.Threshold)

	// bool => C.int
	params.half_size = opts.bool(opts.HalfSize)
	params.four_color_rgb = opts.bool(opts.FourColorRGB)
	params.highlight = C.int(opts.Highlight)
	params.use_auto_wb = opts.bool(opts.UseAutoWb)
	params.use_camera_wb = opts.bool(opts.UseCameraWb)
	params.use_camera_matrix = C.int(opts.UseCameraMatrix)

	params.output_color = C.int(opts.OutputColor)

	// Free me :)
	if opts.OutputProfile != "" {
		params.output_profile = C.CString(opts.OutputProfile)
	}

	if opts.CameraProfile != "" {
		params.camera_profile = C.CString(opts.CameraProfile)
	}

	if opts.BadPixels != "" {
		params.bad_pixels = C.CString(opts.BadPixels)
	}

	if opts.DarkFrame != "" {
		params.dark_frame = C.CString(opts.DarkFrame)
	}

	params.output_bps = C.int(opts.OutputBps)
	params.output_tiff = opts.bool(opts.OutputTiff)
	params.output_flags = C.int(opts.OutputFlags)

	params.user_flip = C.int(opts.UserFlip)
	params.user_qual = C.int(opts.UserQual)
	params.user_black = C.int(opts.UserBlack)
	params.user_cblack = [4]C.int{
		C.int(opts.UserCblack[0]),
		C.int(opts.UserCblack[1]),
		C.int(opts.UserCblack[2]),
		C.int(opts.UserCblack[3]),
	}
	params.user_sat = C.int(opts.UserSat)
	params.med_passes = C.int(opts.MedPasses)
	params.auto_bright_thr = C.float(opts.AutoBrightThr)
	params.adjust_maximum_thr = C.float(opts.AdjustMaximumThr)
	params.no_auto_bright = opts.bool(opts.NoAutoBright)
	params.use_fuji_rotate = opts.bool(opts.UseFujiRotate)
	params.green_matching = opts.bool(opts.GreenMatching)
	params.dcb_iterations = C.int(opts.DcbIterations)
	params.dcb_enhance_fl = opts.bool(opts.DcbEnhanceFl)
	params.fbdd_noiserd = C.int(opts.FbddNoiserd)
	params.exp_correc = opts.bool(opts.ExpCorrect)
	params.exp_shift = C.float(opts.ExpShift)
	params.exp_preser = C.float(opts.ExpPreser)
	params.no_auto_scale = opts.bool(opts.NoAutoScale)
	params.no_interpolation = opts.bool(opts.NoInterpolation)

	return params
}


func (opts *ProcessorOptions) Free(params C.libraw_output_params_t) {
	if opts.OutputProfile != "" {
		C.free(unsafe.Pointer(params.output_profile))
	}

	if opts.CameraProfile != "" {
		C.free(unsafe.Pointer(params.camera_profile))
	}

	if opts.BadPixels != "" {
		C.free(unsafe.Pointer(params.bad_pixels))
	}

	if opts.DarkFrame != "" {
		C.free(unsafe.Pointer(params.dark_frame))
	}


}

// NewProcessorOptions creates a ProcessorOptions struct with the default values from LibRaw
func NewProcessorOptions() ProcessorOptions {
	return ProcessorOptions{
		Greybox:  Box{0, 0, 0, 0},
		Cropbox:  Box{0, 0, 0, 0},
		Aber:     [4]float64{1.0, 1.0, 1.0, 1.0},
		Gamm:     [6]float64{0.45, 4.5, 0.0, 0.0, 0.0, 0.0},
		UserMul:  [4]float32{0.0, 0.0, 0.0, 0.0},
		Bright:   1.0,
		Threshold: 0.0,

		HalfSize:          false,
		FourColorRGB:      false,
		Highlight:         0,
		UseAutoWb:         false,
		UseCameraWb:       false,
		UseCameraMatrix:   1,

		OutputColor: 1,
		OutputProfile: "",
		CameraProfile: "",
		BadPixels: "",
		DarkFrame: "",

		OutputBps:    8,
		OutputTiff:   false,
		OutputFlags:  0,
		UserFlip:     -1,
		UserQual:     -1,
		UserBlack:    -1,
		UserCblack:   [4]int{0, 0, 0, 0},
		UserSat:      -1,

		MedPasses:          0,
		AutoBrightThr:      0.01,
		AdjustMaximumThr:   0.75,
		NoAutoBright:       false,
		UseFujiRotate:      true,
		GreenMatching:      false,
		DcbIterations:      0,
		DcbEnhanceFl:       false,
		FbddNoiserd:        0,
		ExpCorrect:         false,
		ExpShift:           1.0,
		ExpPreser:          0.0,
		NoAutoScale:        false,
		NoInterpolation:    false,
	}
}


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

	proc.params = p.options.Apply(proc.params)
	defer p.options.Free(proc.params)

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
				value := (uint16(dataBytes[i]) << 8) | uint16(dataBytes[i+1])
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

