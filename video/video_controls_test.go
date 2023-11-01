package video

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestControlVisibility(t *testing.T) {
	setup(t)

	widget := NewPlayer()
	widget.SetAutoHideTimer(500 * time.Millisecond) // because 3s is too long for a test
	window := test.NewWindow(widget)

	pipeline := `
    videotestsrc name={{.InputElementName}} !
    videoconvert !
    videoscale !
    video/x-raw,width=320,height=240 !
    videorate name={{.VideoRateElementName}} !
    jpegenc name={{.ImageEncoderElementName}} !
    appsink name={{ .AppSinkElementName }}
    `

	err := widget.SetPipelineFromString(pipeline)
	if err != nil {
		t.Fatal(err)
	}

	window.Resize(fyne.NewSize(800, 600))
	window.ShowAndRun()

	controls := widget.controls
	assert.NotNil(t, controls, "controls should not be nil")

	// check visibility
	assert.True(t, controls.Visible())

	// wait for controls to be hidden
	time.Sleep(1 * time.Second)
	assert.False(t, controls.Visible())
}

func TestShowVideoControls(t *testing.T) {
	setup(t)

	widget := NewPlayer()
	widget.SetAutoHideTimer(500 * time.Millisecond) // because 3s is too long for a test
	window := test.NewWindow(widget)

	pipeline := `
    videotestsrc name={{.InputElementName}} !
    videoconvert !
    videoscale !
    video/x-raw,width=320,height=240 !
    videorate name={{.VideoRateElementName}} !
    videobalance name={{.VideoBalanceElementName}} !
    jpegenc name={{.ImageEncoderElementName}} !
    appsink name={{ .AppSinkElementName }}
    `

	err := widget.SetPipelineFromString(pipeline)
	if err != nil {
		t.Fatal(err)
	}

	window.Resize(fyne.NewSize(800, 600))
	window.ShowAndRun()
	{
		overlay := window.Canvas().Overlays().Top()
		assert.Nil(t, overlay)
	}

	controls := widget.controls

	// press the controls.renderer.controls button, it should show the controls
	controls.renderer.videoControlsButton.Tapped(&fyne.PointEvent{
		Position: fyne.NewPos(0, 0),
	})

	{
		overlay := window.Canvas().Overlays().Top()
		assert.NotNil(t, overlay)
		assert.True(t, overlay.Visible())
	}

}
