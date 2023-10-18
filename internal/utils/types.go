package utils

import (
	"time"

	"fyne.io/fyne/v2"
)

type MediaReader interface {
	Open(uri fyne.URI) error
	Play() error
	Pause() error
	Seek(time.Duration) error
	Duration() (time.Duration, error)
}
