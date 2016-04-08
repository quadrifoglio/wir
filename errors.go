package main

import (
	"errors"
)

var (
	ErrInvalidBackend = errors.New("Invalid backend")
	ErrBackend        = errors.New("Backend error")
)
