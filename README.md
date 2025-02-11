# go-libraw
Go bindign for [LibRaw](https://www.libraw.org/)

# Building
When using MacOS with an M1/M2:
```
brew install libraw
go build .
```

Other OS: (not tested)
1. Install libraw
2. Update the `#cgo` flags to point `libraw`
3. Run `go build .`

