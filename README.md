# go-libraw
Go binding for [LibRaw](https://www.libraw.org/)

## Background
I needed a go binding for LibRaw to convert RAW image formats from Nikon and Canon cameras (.NEF, .CR2, ...) to JPEGs or PNGs.
After doing some searching I only found some older wrappers that have been inactive for a while, so I decided to write my own.

## Install
`go get github.com/seppedelanghe/go-libraw@v0.2.0`

## Options
The LibRaw output params are passed to LibRaw using the `ProcessorOptions` struct. More information about LibRaw Output params can be found [here](https://www.libraw.org/docs/API-datastruct-eng.html#libraw_output_params_t)
The Go struct has some small differences with the LibRaw struct to prevent setting invalid values. 
For example, LibRaw's struct uses 0 and 1 to represent boolean values, in the Go struct I just used the `bool` type to avoid confusion.
A custom struct `Box` is also introduced to avoid setting x, y, w, h values in the wrong order as in C, a `[4]uint` array would be used.

## Building
MacOS:
```
brew install libraw
go build .
```

Ubuntu:
1. Install libraw -> `apt install libraw-dev`
2. Run `go build .`

Other:
1. Install libraw (often called `libraw-dev`)
2. (optional) Update the `#cgo` flags to point the correct directory for `libraw` 
3. Run `go build .`

### Tested on:
- MacOS 13
- Ubuntu 24.04 ARM
- Ubuntu 24.04 x64

## Tests
Running tests can be done using `go test -v .`

## Example usage
```
const pathToRawFile = "./dir/file.NEF"
processor := libraw.NewProcessor(libraw.NewProcessorOptions())
img, metadata, err := processor.ProcessRaw(pathToRawFile)
// handle err...
```

For a full example see: `cmd/example.go`


## Credits
- Existing golibraw wrapper by enricod: https://github.com/enricod/golibraw
