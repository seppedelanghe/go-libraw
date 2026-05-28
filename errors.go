package golibraw

import "errors"

var ErrLibrawInit = errors.New("failed to initialaze libraw")
var ErrNoPreview = errors.New("libraw: image has no preview")
