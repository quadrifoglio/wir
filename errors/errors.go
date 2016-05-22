package errors

import (
	"errors"
)

var (
	NotFound            = errors.New("Not Found")
	BadRequest          = errors.New("Bad Request")
	InvalidImageType    = errors.New("Invalid Image Type")
	InvalidURL          = errors.New("Invalid URL")
	UnsupportedProto    = errors.New("Unsupported Protocol")
	ImageNotFound       = errors.New("Image Not Found")
	InvalidMachineState = errors.New("Invalid Machine State")
	StartFailed         = errors.New("Machine Start Failed")
	KillFailed          = errors.New("Process kill failed")
)
