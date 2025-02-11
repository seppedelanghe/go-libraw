# go-libraw
Go bindign for [LibRaw](https://www.libraw.org/)

## Background
I needed a go binding for LibRaw to convert RAW image formats from Nikon and Canon cameras (.NEF, .CR2, ...) to JPEGs or PNGs.
After doing some searching I only found some older wrappers that have been inactive for a while, so I decided to write my own.

I use MacOS with a M1 Pro chip, so this is only tested (for now) on ARM.

## Building
When using MacOS with an M1/M2:
```
brew install libraw
go build .
```

Other OS: (not tested)
1. Install libraw
2. Update the `#cgo` flags to point `libraw`
3. Run `go build .`

## Example usage
```
const pathToRawFile = "./dir/file.NEF"
processor := libraw.NewProcessor(libraw.ProcessorOptions{})
img, metadata, err := processor.ProcessRaw(pathToRawFile)
// handle err...
```

For a full example see: `cmd/example.go`


## Credits
- Existing golibraw wrapper by enricod: https://github.com/enricod/golibraw
