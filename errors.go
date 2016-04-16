package main

import (
	"errors"
)

var (
	ErrInvalidBackend  = errors.New("Invalid backend")
	ErrBackend         = errors.New("Backend error")
	ErrImageNotFound   = errors.New("Image not found")
	ErrVmNotFound      = errors.New("Vm not found")
	ErrNoAttrs         = errors.New("No vm attributes")
	ErrInvalidAttrType = errors.New("Invalid attribute type")
	ErrProcessNotFound = errors.New("Vm process not found")
	ErrStart           = errors.New("Can not start vm")
	ErrKill            = errors.New("Can not kill vm process")
	ErrRunning         = errors.New("Vm already running")
	ErrNotRunning      = errors.New("Vm not running")
)
