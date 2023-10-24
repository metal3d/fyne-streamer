package video

import (
	"context"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/test"
	"github.com/go-gst/go-gst/gst"
	"github.com/metal3d/fyne-streamer/internal/utils"
	"github.com/stretchr/testify/assert"
)

func init() {
	utils.GstreamerInit()
}

func TestCreateAViewer(t *testing.T) {
	viewer := NewViewer()
	_ = test.WidgetRenderer(viewer)
	pipeline := viewer.Pipeline()
	assert.Nil(t, pipeline)
}

func TestCheckMandatoryElements(t *testing.T) {
	video := NewViewer()
	_ = test.WidgetRenderer(video)
	err := video.SetPipelineFromString(`
    videotestsrc ! fakesink`)
	assert.NotNil(t, err)
}

func TestCreateVideo(t *testing.T) {
	video := NewViewer()
	_ = test.WidgetRenderer(video)
	err := video.SetPipelineFromString(`
    videotestsrc name={{.InputElementName}} ! 
    videoconvert ! 
    videoscale !
    video/x-raw,width=320,height=240 !
    videorate name={{.VideoRateElementName}} !
    jpegenc name={{.ImageEncoderElementName}} ! 
    appsink name={{ .AppSinkElementName }}`)
	assert.Nil(t, err)

	// pipeline should not be nil
	pipeline := video.Pipeline()
	assert.NotNil(t, pipeline)

	// state shuld be Null
	state := pipeline.GetCurrentState()
	assert.Equal(t, state, gst.StateNull)

	video.Play()
	state = pipeline.GetCurrentState()
	assert.Equal(t, gst.StatePlaying, state)
	assert.True(t, video.IsPlaying())

	// check that the video can be paused
	time.Sleep(1 * time.Second)
	video.Pause()
	state = pipeline.GetCurrentState()
	assert.Equal(t, gst.StatePaused, state)
	assert.False(t, video.IsPlaying())

	// ensure that the widget is 320x240
	assert.Equal(t, video.VideoSize().Width, float32(320))
	assert.Equal(t, video.VideoSize().Height, float32(240))
}

func TestOpeningFile(t *testing.T) {

	video := NewViewer()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	preRollReached := false
	newFrameReached := 0
	hasEOSReached := false

	video.SetOnEOS(func() {
		t.Log("EOS")
		hasEOSReached = true
		cancel()
	})

	video.SetOnPreRoll(func() {
		preRollReached = true
	})

	video.SetOnNewFrame(func(at time.Duration) {
		newFrameReached += 1
	})

	_ = test.WidgetRenderer(video)
	err := video.Open(storage.NewFileURI(_testVideoFile))
	assert.Nil(t, err)

	// pipeline should not be nil
	pipeline := video.Pipeline()
	assert.NotNil(t, pipeline)

	// state shuld be Null
	state := pipeline.GetCurrentState()
	assert.Equal(t, gst.StateNull, state)

	err = video.Play()
	assert.Nil(t, err)
	<-ctx.Done()

	assert.True(t, hasEOSReached)
	assert.True(t, preRollReached)
	assert.True(t, newFrameReached > 0)
}

func TestBalance(t *testing.T) {
	video := NewViewer()
	_ = test.WidgetRenderer(video)
	video.SetPipelineFromString(`
    videotestsrc name={{.InputElementName}} !
    videoconvert !
    videoscale !
    video/x-raw,width=320,height=240 !
    videorate name={{.VideoRateElementName}} !
    videobalance name={{.VideoBalanceElementName}} !
    jpegenc name={{.ImageEncoderElementName}} !
    appsink name={{ .AppSinkElementName }}`)

	video.Play()
	contrast, brightness, hue, saturation := video.GetVideoBalance()
	assert.Equal(t, float64(1), contrast)
	assert.Equal(t, float64(0), brightness)
	assert.Equal(t, float64(0), hue)
	assert.Equal(t, float64(1), saturation)

	video.SetContrast(0)
	contrast = video.GetContrast()
	assert.Equal(t, float64(0), contrast)

	video.SetBrightness(1)
	brightness = video.GetBrightness()
	assert.Equal(t, float64(1), brightness)

	video.SetHue(1)
	hue = video.GetHue()
	assert.Equal(t, float64(1), hue)

	video.SetSaturation(0)
	saturation = video.GetSaturation()
	assert.Equal(t, float64(0), saturation)
}

func TestFullScreenMode(t *testing.T) {
	video := NewViewer()
	win := test.NewWindow(video)
	win.Resize(fyne.NewSize(320, 240))
	win.Show()

	video.SetPipelineFromString(`
    videotestsrc name={{.InputElementName}} !
    videoconvert !
    videoscale !
    video/x-raw,width=320,height=240 !
    videorate name={{.VideoRateElementName}} !
    jpegenc name={{.ImageEncoderElementName}} !
    appsink name={{ .AppSinkElementName }}`)
	video.Play()

	video.SetFullScreen(true)
	assert.NotNil(t, video.fullscreenWindow)
	assert.True(t, video.fullscreenWindow.FullScreen())

	video.SetFullScreen(false)
}

func TestSeek(t *testing.T) {
	video := NewViewer()
	_ = test.WidgetRenderer(video)
	err := video.Open(storage.NewFileURI(_testVideoFile))
	assert.Nil(t, err)

	err = video.Play()
	assert.Nil(t, err, "cannot play")

	// wait ready state
	context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	video.SetOnNewFrame(func(at time.Duration) {
		cancel()
	})
	<-context.Done()

	// seek to 2 seconds
	time.Sleep(500 * time.Millisecond)
	err = video.Seek(2 * time.Second)
	assert.Nil(t, err, "error while seeking position")
}
