package api

import (
	"github.com/quadrifoglio/wir/shared"
)

type Image interface {
	Create(info shared.ImageInfo) error
	Info() shared.ImageInfo
	Delete() error
}
