package utils

import (
	"time"

	"fyne.io/fyne/v2"
)

type MediaOpener interface {
	Open(uri fyne.URI) error
}

type MediaControl interface {
	Play() error
	Pause() error
}

type MediaSeeker interface {
	Seek(time.Duration) error
}

type MediaDuration interface {
	Duration() (time.Duration, error)
}
