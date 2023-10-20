package video

import (
	"context"
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

const viewerVideoTimeout = 5 * time.Second

func init() {
	utils.GstreamerInit()
}

func _createViewerTestingVideo() {
	pipeline, err := gst.NewPipelineFromString(`
        videotestsrc !
        videoconvert !
        videoscale !
        video/x-raw,width=320,height=240 !
        vp8enc !
        webmmux !
        filesink location=/tmp/test.webm`)
	if err != nil {
		panic(err)
	}
	pipeline.SetState(gst.StatePlaying)
	time.Sleep(viewerVideoTimeout)
	pipeline.SetState(gst.StateNull)
	pipeline.Clear()
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
	assert.Equal(t, state, gst.StatePlaying)
	assert.True(t, video.IsPlaying())

	// ensure that the widget is 320x240
	assert.Equal(t, video.VideoSize().Width, float32(320))
	assert.Equal(t, video.VideoSize().Height, float32(240))
}

func TestOpeningFile(t *testing.T) {
	_createViewerTestingVideo()
	defer os.Remove("/tmp/test.webm")

	video := NewViewer()
	ctx, cancel := context.WithTimeout(context.Background(), viewerVideoTimeout+1*time.Second)

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
	err := video.Open(storage.NewFileURI("/tmp/test.webm"))
	assert.Nil(t, err)

	// pipeline should not be nil
	pipeline := video.Pipeline()
	assert.NotNil(t, pipeline)

	// state shuld be Null
	state := pipeline.GetCurrentState()
	assert.Equal(t, state, gst.StateNull)

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
