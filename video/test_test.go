package video

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/metal3d/fyne-streamer/internal/utils"
)

const (
	_testVideoFile = "./test-files/testvideo.ogv"
)

var logBuffer bytes.Buffer

func setup(t *testing.T) {
	utils.GstreamerInit()
	log.SetOutput(&logBuffer)
	t.Cleanup(func() {
		logBuffer.Reset()
		log.SetOutput(os.Stderr)
	})
}
