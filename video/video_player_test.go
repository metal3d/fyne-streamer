package video

import (
	"fmt"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/test"
	"github.com/go-gst/go-gst/gst"
	"github.com/metal3d/fyne-streamer/internal/utils"
	"github.com/stretchr/testify/assert"
)

const (
	playerVideoTimeout = 5 * time.Second
	testVideoPath      = "/tmp/test.webm"
)

func init() {
	utils.GstreamerInit()
}

func _createPlayerTestingVideo() {
	pipeline, err := gst.NewPipelineFromString(fmt.Sprintf(`
        videotestsrc !
        videoconvert !
        videoscale !
        video/x-raw,width=320,height=240 !
        vp8enc !
        webmmux !
        filesink location=%q`, testVideoPath))

	if err != nil {
		panic(err)
	}
	pipeline.SetState(gst.StatePlaying)
	time.Sleep(viewerVideoTimeout)
	pipeline.SetState(gst.StateNull)
	pipeline.Clear()
}

func TestCreatePlayer(t *testing.T) {
	player := NewPlayer()
	_ = test.WidgetRenderer(player)
	pipeline := player.Pipeline()
	assert.Nil(t, pipeline)
}

func TestPlayerCursor(t *testing.T) {
	_createPlayerTestingVideo()
	defer os.Remove("/tmp/test.webm")

	player := NewPlayer()
	window := test.NewWindow(player)
	window.Resize(fyne.NewSize(320, 240))
	window.Show()

	err := player.Open(storage.NewFileURI(testVideoPath))
	assert.Nil(t, err)

	v := player.controls.renderer.cursor.Value
	assert.Equal(t, float64(0), v)

	player.Play()
	time.Sleep(1 * time.Second)

	v = player.controls.renderer.cursor.Value
	assert.True(t, v > 0)
}
