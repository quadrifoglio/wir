package errors

import (
	"errors"
)

var (
	NotFound         = errors.New("Not Found")
	BadRequest       = errors.New("Bad Request")
	InvalidImageType = errors.New("Invalid Image Type")
	InvalidURL       = errors.New("Invalid URL")
	UnsupportedProto = errors.New("Unsupported Protocol")
)
