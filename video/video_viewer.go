package video

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/go-gst/go-gst/gst"
	gstApp "github.com/go-gst/go-gst/gst/app"
	streamer "github.com/metal3d/fyne-streamer"
	"github.com/metal3d/fyne-streamer/internal/utils"
)

var _ fyne.Widget = (*Viewer)(nil)
var _ utils.MediaControl = (*Viewer)(nil)
var _ utils.MediaDuration = (*Viewer)(nil)
var _ utils.MediaOpener = (*Viewer)(nil)
var _ utils.MediaSeeker = (*Viewer)(nil)

// Viewer widget is a simple video player with no controls to display.
// This is a base widget to only read a video or that can be extended to create a video player with controls.
type Viewer struct {
	widget.BaseWidget
	pipeline         *gst.Pipeline
	appSink          *gstApp.Sink
	onNewFrame       func(time.Duration)
	onPreRoll        func()
	onEOS            func()
	onPaused         func()
	onStartPlaying   func()
	onTitle          func(string)
	rate             int
	imageQuality     int
	width            int
	height           int
	duration         time.Duration
	frame            *canvas.Image
	fullscreenWindow fyne.Window
	currentWindow    fyne.Window
	bus              *gst.Bus

	// currentWindowFinder is a function that returns the current window of the
	// widget. It is used to find the current window when the widget is in
	// fullscreen mode.
	// This is necessary because the current widget can be composed in another.
	// Use setCurrentWindowFinder to set this function.
	currentWindowFinder func() fyne.Window

	// originalViewerWidget is the object that is really displayed in the window.
	// This is important to get it for the fullscreen mode.
	originalViewerWidget fyne.CanvasObject
}

// NewViewer creates a new video viewer widget.
func NewViewer() *Viewer {
	v := CreateBaseVideoViewer()
	v.ExtendBaseWidget(v)
	return v
}

// CreateRenderer creates a renderer for the video widget.
//
// Implements: fyne.Widget
func (v *Viewer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(v.frame)
}

// CurrentPosition returns the current position of the stream in time.
func (v *Viewer) CurrentPosition() (time.Duration, error) {
	if v.pipeline == nil {
		return 0, streamer.ErrNoPipeline
	}
	ok, pos := v.pipeline.QueryPosition(gst.FormatTime)
	if !ok {
		return 0, streamer.ErrPositionUnseekable
	}
	return time.Duration(float64(pos)), nil
}

// Duration returns the duration of the stream if possible.
func (v *Viewer) Duration() (time.Duration, error) {
	if v.pipeline == nil {
		return 0, streamer.ErrNoPipeline
	}
	if v.duration != 0 {
		return v.duration, nil
	}
	ok, duration := v.pipeline.QueryDuration(gst.FormatTime)
	if !ok {
		return 0, streamer.ErrNoDuration
	}
	v.duration = time.Duration(float64(duration))
	return v.duration, nil
}

// ExtendBaseWidget overrides the ExtendBaseWidget method of the BaseWidget.
// It is used to set the currentWindowFinder function and the object that
// is really displayed in the window (to ensure that fullscreen
// will use the right object).
func (v *Viewer) ExtendBaseWidget(w fyne.Widget) {
	v.BaseWidget.ExtendBaseWidget(w)
	v.setCurrentWindowFinder(w)
	v.originalViewerWidget = w
}

// Frame returns the canvas.Image that is used to display the video.
func (v *Viewer) Frame() *canvas.Image {
	return v.frame
}

// GetBrightness returns the brightness of the video.
func (v *Viewer) GetBrightness() float64 {
	if v.pipeline == nil {
		return 0
	}
	videoBalanceElement, err := v.pipeline.GetElementByName(streamer.VideoBalanceElementName)
	if err != nil {
		fyne.LogError("Failed to find the video balance element", err)
		return 0
	}
	brightness, err := videoBalanceElement.GetProperty("brightness")
	if err != nil {
		fyne.LogError("Failed to get the brightness property", err)
		return 0
	}
	return brightness.(float64)
}

// GetContrast returns the contrast of the video.
func (v *Viewer) GetContrast() float64 {
	if v.pipeline == nil {
		return 0
	}
	videoBalanceElement, err := v.pipeline.GetElementByName(streamer.VideoBalanceElementName)
	if err != nil {
		fyne.LogError("Failed to find the video balance element", err)
		return 0
	}
	contrast, err := videoBalanceElement.GetProperty("contrast")
	if err != nil {
		fyne.LogError("Failed to get the contrast property", err)
		return 0
	}
	return contrast.(float64)
}

// GetHue returns the hue of the video.
func (v *Viewer) GetHue() float64 {
	if v.pipeline == nil {
		return 0
	}
	videoBalanceElement, err := v.pipeline.GetElementByName(streamer.VideoBalanceElementName)
	if err != nil {
		fyne.LogError("Failed to find the video balance element", err)
		return 0
	}
	hue, err := videoBalanceElement.GetProperty("hue")
	if err != nil {
		fyne.LogError("Failed to get the hue property", err)
		return 0
	}
	return hue.(float64)
}

// GetSaturation returns the saturation of the video.
func (v *Viewer) GetSaturation() float64 {
	if v.pipeline == nil {
		return 0
	}
	videoBalanceElement, err := v.pipeline.GetElementByName(streamer.VideoBalanceElementName)
	if err != nil {
		fyne.LogError("Failed to find the video balance element", err)
		return 0
	}
	saturation, err := videoBalanceElement.GetProperty("saturation")
	if err != nil {
		fyne.LogError("Failed to get the saturation property", err)
		return 0
	}
	return saturation.(float64)
}

// GetVideoBalance returns the contrast, brightness, hue and saturation of the video.
func (v *Viewer) GetVideoBalance() (float64, float64, float64, float64) {
	return v.GetContrast(), v.GetBrightness(), v.GetHue(), v.GetSaturation()
}

// IsMuted returns true if the audio is muted.
func (v *Viewer) IsMuted() bool {
	if v.pipeline == nil {
		return false
	}
	volumeElement, err := v.pipeline.GetElementByName(streamer.VolumeElementName)
	if err != nil {
		fyne.LogError("Failed to find the volume element", err)
		return false
	}

	isMuted, err := volumeElement.GetProperty("mute")
	if err != nil {
		fyne.LogError("Failed to get the mute property", err)
		return false
	}
	return isMuted.(bool)
}

// IsPlaying returns true if the pipeline is in the playing state.
func (v *Viewer) IsPlaying() bool {
	if v.pipeline == nil {
		return false
	}
	return v.pipeline.GetCurrentState() == gst.StatePlaying
}

// Mute the audio.
func (v *Viewer) Mute() {
	if v.pipeline == nil {
		return
	}
	volumeElement, err := v.pipeline.GetElementByName(streamer.VolumeElementName)
	if err != nil {
		fyne.LogError("Failed to find the volume element", err)
		return
	}
	volumeElement.SetProperty("mute", true)
}

func (v *Viewer) OnStartPlaying(f func()) {
	v.onStartPlaying = f
}

// Pause the stream if the pipeline is not nil.
func (v *Viewer) Pause() error {
	if v.pipeline == nil {
		return streamer.ErrNoPipeline
	}
	defer func() {
		if v.onPaused != nil {
			v.onPaused()
		}
	}()
	if err := v.SetState(gst.StatePaused); err != nil {
		return err
	}
	// BUG: This fix some issues with the pipeline that paused then becomes crazy when we start it again. It has the effect to sync the pipeline.
	if pos, err := v.CurrentPosition(); err == nil {
		v.Seek(pos)
	}
	return nil
}

// Pipeline returns the gstreamer pipeline.
func (v *Viewer) Pipeline() *gst.Pipeline {
	return v.pipeline
}

// Play the stream if the pipeline is not nil.
func (v *Viewer) Play() error {
	if v.pipeline == nil {
		return streamer.ErrNoPipeline
	}
	defer func() {
		if v.onStartPlaying != nil {
			v.onStartPlaying()
		}
	}()

	err := v.SetState(gst.StatePlaying)
	return err
}

// Seek the position to "pos" Nanoseconds. Set the playing stream to this time position.
// If the element or the pipeline cannot be seekable, the operation is cancelled.
func (v *Viewer) Seek(pos time.Duration) error {
	if v.pipeline == nil {
		return streamer.ErrNoPipeline
	}
	query := gst.NewSeekingQuery(gst.FormatTime)
	if !v.pipeline.Query(query) {
		return streamer.ErrSeekUnsupported
	}

	defer v.resync()
	done := v.pipeline.SeekTime(pos, gst.SeekFlagFlush)
	if !done {
		return streamer.ErrSeekFailed
	}
	v.appSink.Element.SyncStateWithParent()
	return nil
}

// SetBrightness sets the brightness of the video.
func (v *Viewer) SetBrightness(brightness float64) {
	if v.pipeline == nil {
		return
	}
	if brightness < -1 || brightness > 1 {
		return
	}
	videoBalanceElement, err := v.pipeline.GetElementByName(streamer.VideoBalanceElementName)
	if err != nil {
		fyne.LogError("Failed to find the video balance element", err)
		return
	}
	videoBalanceElement.SetProperty("brightness", brightness)
}

// SetContrast sets the contrast of the video.
func (v *Viewer) SetContrast(contrast float64) {
	if v.pipeline == nil {
		return
	}
	if contrast < 0 || contrast > 2 {
		return
	}
	videoBalanceElement, err := v.pipeline.GetElementByName(streamer.VideoBalanceElementName)
	if err != nil {
		fyne.LogError("Failed to find the video balance element", err)
		return
	}
	videoBalanceElement.SetProperty("contrast", contrast)
}

// SetFillMode sets the fill mode of the image.
func (v *Viewer) SetFillMode(mode canvas.ImageFill) {
	v.frame.FillMode = mode

}

// SetFullScreen sets the video widget to fullscreen mode or not.
func (v *Viewer) SetFullScreen(state bool) {
	if state {
		v.fullscreenOn()
	} else {
		v.fullscreenOff()
	}
}

// SetHue sets the hue of the video.
func (v *Viewer) SetHue(hue float64) {
	if v.pipeline == nil {
		return
	}
	if hue < -1 || hue > 1 {
		return
	}
	videoBalanceElement, err := v.pipeline.GetElementByName(streamer.VideoBalanceElementName)
	if err != nil {
		fyne.LogError("Failed to find the video balance element", err)
		return
	}
	videoBalanceElement.SetProperty("hue", hue)
}

// SetMaxRate sets the max-rate property of the videorate element.
// It is used to limit the framerate of the video.
// Note that if the rate value is too high, the Video element will fix it automatically.
func (v *Viewer) SetMaxRate(rate int) error {

	if v.pipeline == nil {
		return streamer.ErrNoPipeline
	}
	rateElement, err := v.pipeline.GetElementByName(streamer.VideoRateElementName)
	if err != nil {
		return fmt.Errorf("failed to find the video rate element: %w", err)
	}

	if rateElement == nil {
		return fmt.Errorf("no video rate element in the pipeline")
	}

	duration := time.Duration(float64(time.Second) / float64(rate))
	if err := rateElement.SetProperty("max-rate", int(duration)); err != nil {
		return fmt.Errorf("failed to set max-rate property: %w", err)
	}

	// try to set the max-lateness of the appsink
	if v.appSink != nil {
		if err := v.appSink.SetProperty("max-lateness", int64(duration)); err != nil {
			return fmt.Errorf("failed to set max-lateness property: %w", err)
		}
	}

	v.rate = rate
	return nil
}

// SetOnEOS set the function to call when EOS is reached in the pipeline. E.g. when the// video ends.
func (v *Viewer) SetOnEOS(f func()) {
	v.onEOS = f
}

// SetOnNewFrame set the function that is called when a new frame is available and presented to the view. The position is set as time.Duration to the function.
func (v *Viewer) SetOnNewFrame(f func(time.Duration)) {
	v.onNewFrame = f
}

// SetOnPaused set the function that is called when the pipeline is paused.
func (v *Viewer) SetOnPaused(f func()) {
	v.onPaused = f
}

// SetOnPreRoll set the function that is called when the pipeline is prerolling. At this time, you must be able to get the video size and duration.
func (v *Viewer) SetOnPreRoll(f func()) {
	v.onPreRoll = f
}

func (v *Viewer) SetOnTitle(f func(string)) {
	v.onTitle = f
}

// SetQuality of the jpeg encoder. If que quality is not between 0 and 100, nothing is done.
func (v *Viewer) SetQuality(q int) error {
	if v.pipeline == nil {
		return streamer.ErrNoPipeline
	}
	imageEncoderElement, err := v.pipeline.GetElementByName(streamer.ImageEncoderElementName)
	if err != nil {
		return fmt.Errorf("failed to find the image encoder element: %w", err)
	}

	if q < 0 || q > 100 {
		return fmt.Errorf("the quality should be 0 < s < 100, given value %v", q)
	}

	pluginname := imageEncoderElement.GetFactory().GetName()
	switch pluginname {
	case "jpegenc":
		if err := imageEncoderElement.SetProperty("quality", q); err != nil {
			return fmt.Errorf("failed to set quality property: %w", err)
		}
	case "pngenc":
		compressionLevel := 9 - (q * 9 / 100)
		if err := imageEncoderElement.SetProperty("compression-level", compressionLevel); err != nil {
			// compression-level 0 to 9, convert the quality to compression-level
			// 0 is the best quality, 9 is the worst
			return fmt.Errorf("failed to set quality property: %w", err)
		}
	default:
		return fmt.Errorf("the image encoder element is not jpegenc or pngenc, it is %s", pluginname)
	}

	v.imageQuality = q
	return nil
}

// SetSaturation sets the saturation of the video.
func (v *Viewer) SetSaturation(saturation float64) {
	if v.pipeline == nil {
		return
	}
	if saturation < 0 || saturation > 2 {
		return
	}
	videoBalanceElement, err := v.pipeline.GetElementByName(streamer.VideoBalanceElementName)
	if err != nil {
		fyne.LogError("failed to find the video balance element", err)
		return
	}
	videoBalanceElement.SetProperty("saturation", saturation)
}

// SetScaleMode sets the scale mode of the image. It's not recommended to use other
// mode than canvas.ImageScaleFastest because it can be very slow.
func (v *Viewer) SetScaleMode(mode canvas.ImageScale) {
	v.frame.ScaleMode = mode
}

// SetState sets the state of the pipeline to the given state.
func (v *Viewer) SetState(state gst.State) error {
	defer v.resync()
	return v.setState(state)
}

// SetVolume sets the volume of the audio.
func (v *Viewer) SetVolume(volume float64) {
	if v.pipeline == nil {
		return
	}
	if volume < 0 || volume > 1 {
		return
	}
	volumeElement, err := v.pipeline.GetElementByName(streamer.VolumeElementName)
	if err != nil {
		fyne.LogError("failed to find the volume element", err)
		return
	}
	volumeElement.SetProperty("volume", volume)
}

// ToggleMute mutes or unmutes the audio.
func (v *Viewer) ToggleMute() {
	if v.IsMuted() {
		v.Unmute()
	} else {
		v.Mute()
	}
}

// Unmute the audio.
func (v *Viewer) Unmute() {
	if v.pipeline == nil {
		return
	}
	volumeElement, err := v.pipeline.GetElementByName(streamer.VolumeElementName)
	if err != nil {
		fyne.LogError("failed to find the volume element", err)
		return
	}
	volumeElement.SetProperty("mute", false)
}

// VideoSize returns the size of the video (resolution in pixels).
func (v *Viewer) VideoSize() fyne.Size {
	return fyne.NewSize(float32(v.width), float32(v.height))
}

// CreateBaseVideoViewer returns a new video widget without the base widget.
// It is useful to create various video widgets with the same base.
// It's MANDATORY to use it to create viewers.
func CreateBaseVideoViewer() *Viewer {
	utils.GstreamerInit()
	v := &Viewer{
		frame:        canvas.NewImageFromResource(nil),
		rate:         30,
		imageQuality: 85,
	}

	v.SetFillMode(canvas.ImageFillContain)
	v.SetScaleMode(canvas.ImageScaleFastest)
	return v
}
