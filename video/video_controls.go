package video

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/go-gst/go-gst/gst"
	streamer "github.com/metal3d/fyne-streamer"
)

// autoHideDuration is the default duration of the auto hide of the controls.
const autoHideDuration = 2 * time.Second

var _ fyne.Widget = (*VideoControls)(nil)
var _ fyne.WidgetRenderer = (*videoControlsRenderer)(nil)

// VideoControls is the widget that displays the video controls (play, pause, fullscreen, etc.).
type VideoControls struct {
	widget.BaseWidget
	viewer   *Viewer
	timeStep time.Duration
	renderer *videoControlsRenderer
	onTapped func()
}

// NewVideoControls creates a new video controls widget. It is used to control the video viewer.
// It is use in the Player widget.
func NewVideoControls(viewer *Viewer) *VideoControls {
	vc := &VideoControls{
		viewer:   viewer,
		timeStep: autoHideDuration,
	}
	vc.ExtendBaseWidget(vc)
	return vc
}

// CreateRenderer returns a new videoControlsRenderer.
//
// Implements fyne.Widget.
func (vc *VideoControls) CreateRenderer() fyne.WidgetRenderer {
	r := newVideoControlsRenderer(vc)
	vc.renderer = r
	return r
}

// Refresh the video controls.
func (vc *VideoControls) Refresh() {
	if vc.renderer == nil {
		return
	}
	vc.renderer.Refresh()
}

// SetCursorAt sets the cursor at the given position.
func (vc *VideoControls) SetCursorAt(pos time.Duration) {
	vc.renderer.manualSeeked = false
	vc.renderer.cursor.SetValue(float64(pos.Milliseconds()))
	vc.renderer.currentTime = pos
	vc.renderer.manualSeeked = true
	vc.Refresh()
}

// SetDuration sets the slider max value to the given duration.
func (vc *VideoControls) SetDuration(d time.Duration) {
	vc.renderer.cursor.Max = float64(d.Milliseconds())
	vc.renderer.totalTime = d
	vc.Refresh()
}

// videoControlsRenderer is the renderer of the video controls widget. This is the widget that
// displays the buttons, sliders, background, etc.
type videoControlsRenderer struct {
	parent           *VideoControls
	playbutton       *widget.Button
	muteButton       *widget.Button
	fullscreenButton *widget.Button
	timeText         *widget.Label
	currentTime      time.Duration
	totalTime        time.Duration
	cursor           *widget.Slider
	controls         *fyne.Container
	background       *canvas.Rectangle
	isFullScreen     bool
	manualSeeked     bool
}

// newVideoControlsRenderer creates a new videoControlsRenderer.
func newVideoControlsRenderer(parent *VideoControls) *videoControlsRenderer {
	renderer := &videoControlsRenderer{
		parent:       parent,
		manualSeeked: true,
	}

	timeText := widget.NewLabel("00:00:00 / 00:00:00")
	timeText.Alignment = fyne.TextAlignCenter

	fullscreenButton := renderer.createFullScreenButton()
	playbutton := renderer.createPlayButton()
	cursor := renderer.cratePositionCursor()
	volumeSlider := renderer.createVolumeSlider()
	backToZeroButton := renderer.createBackToZeroButton()
	stepForwardButton := renderer.createStepForwardButton()
	stepBackwardButton := renderer.createStepBackwardButton()
	volumeMuteButton := renderer.createVolumeMuteButton()

	videoControlsButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), renderer.showVideoControls)
	videoControlsButton.Importance = widget.LowImportance

	controls := container.NewBorder(
		timeText, // the timer (time position / duration)
		cursor,   // the cursor to navigate in the video
		nil, nil,
		container.NewCenter(
			container.NewHBox( // the controls
				backToZeroButton,
				stepBackwardButton,
				playbutton,
				stepForwardButton,
				fullscreenButton,
				volumeSlider,
				volumeMuteButton,
				videoControlsButton,
			),
		),
	)

	// build a semi transparent background based on the theme.BackgroundColor()
	r, b, g, _ := theme.BackgroundColor().RGBA()
	var alpha uint32
	if fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantDark {
		alpha = 0xAF
	} else {
		alpha = 0x9F
	}

	background := canvas.NewRectangle(color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(alpha)})
	background.CornerRadius = theme.InputRadiusSize() * 4

	// register the controls elements
	renderer.controls = controls
	renderer.playbutton = playbutton
	renderer.fullscreenButton = fullscreenButton
	renderer.timeText = timeText
	renderer.cursor = cursor
	renderer.background = background
	renderer.controls = controls
	renderer.muteButton = volumeMuteButton

	return renderer
}

// Destroy is an internal function from the fyne.WidgetRenderer interface.
func (v *videoControlsRenderer) Destroy() { /* no op */ }

// Layout implements the fyne.WidgetRenderer interface. It places the controls at the bottom of the video widget.
func (v *videoControlsRenderer) Layout(size fyne.Size) {
	if v.parent == nil || v.parent.viewer == nil {
		return
	}
	if v.controls == nil {
		log.Println("v.controls is nil")
		return
	}
	v.parent.viewer.Frame().Resize(size)
	v.parent.viewer.Frame().Move(fyne.NewPos(0, 0))

	// the size should be 0.8 of the width
	offsetRight := (size.Width - size.Width*0.8) / 2
	v.controls.Resize(fyne.NewSize(size.Width*.8, v.controls.MinSize().Height))
	// move the controls to the bottom
	upPadding := theme.Padding() * 3
	v.controls.Move(
		fyne.NewPos(
			offsetRight,
			size.Height-v.controls.MinSize().Height-upPadding,
		),
	)

	v.background.Resize(fyne.NewSize(size.Width*.8, v.controls.MinSize().Height+theme.InputRadiusSize()))
	v.background.Move(fyne.NewPos(
		offsetRight,
		size.Height-v.controls.MinSize().Height-upPadding,
	))
}

// MinSize implements the fyne.WidgetRenderer interface. It returns the minimum size of the controls.
func (v *videoControlsRenderer) MinSize() fyne.Size {
	var w float32
	h := v.controls.MinSize().Height + theme.InputRadiusSize()
	for _, o := range v.Objects() {
		if o == nil {
			continue
		}
		if o.MinSize().Width > w {
			w += o.MinSize().Width
		}
	}
	return fyne.NewSize(w, h)
}

// Objects implements the fyne.WidgetRenderer interface. It returns the controls objects.
func (v *videoControlsRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		v.background,
		v.controls,
	}
}

// Refresh implements the fyne.WidgetRenderer interface. It refreshes the controls.
func (v *videoControlsRenderer) Refresh() {
	go v.parent.viewer.Frame().Refresh()
	// convert the time.Duration to a string using TimeFormat
	currentTime := time.Time{}.Add(v.currentTime).Format(streamer.TimeFormat)
	totalTime := time.Time{}.Add(v.totalTime).Format(streamer.TimeFormat)
	v.timeText.SetText(fmt.Sprintf("%s / %s", currentTime, totalTime))
	v.manualSeeked = false
	v.cursor.Max = float64(v.totalTime.Milliseconds())
	v.cursor.SetValue(float64(v.currentTime.Milliseconds()))
	v.manualSeeked = true

	go time.AfterFunc(100*time.Millisecond, func() { // TODO: we need to wait for the state to be updated
		if v.parent.viewer.IsPlaying() {
			v.playbutton.SetIcon(theme.MediaPauseIcon())
		} else {
			v.playbutton.SetIcon(theme.MediaPlayIcon())
		}
	})

	v.controls.Refresh()
	v.background.Refresh()
}

func (v *videoControlsRenderer) cratePositionCursor() *widget.Slider {
	cursor := widget.NewSlider(0, 100)
	cursor.OnChanged = func(value float64) {
		if !v.manualSeeked {
			return
		}
		v.parent.viewer.Seek(time.Duration(value) * time.Millisecond)
		if v.parent.onTapped == nil {
			return
		}
		v.parent.onTapped()
	}
	cursor.OnChangeEnded = func(value float64) {
		if !v.manualSeeked {
			return
		}
		if v.parent.viewer.IsPlaying() {
			return
		}
		im, ret := v.parent.viewer.getCurrentFrame(v.parent.viewer.appSink, true)
		if ret != gst.FlowOK {
			log.Println("error getting frame", ret)
			return
		}
		v.parent.viewer.Frame().Image = im
		v.parent.viewer.Frame().Refresh()
		if v.parent.onTapped != nil {
			v.parent.onTapped()
		}
	}

	return cursor
}

func (v *videoControlsRenderer) createBackToZeroButton() *widget.Button {
	backToZeroButton := widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), func() {
		v.parent.viewer.Seek(0)
		if v.parent.onTapped == nil {
			return
		}
		v.parent.onTapped()
	})
	backToZeroButton.Importance = widget.LowImportance
	return backToZeroButton
}

func (v *videoControlsRenderer) createFullScreenButton() *widget.Button {
	fullscreenButton := widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), func() {
		v.isFullScreen = !v.isFullScreen
		v.parent.viewer.SetFullScreen(v.isFullScreen)
	})
	fullscreenButton.Importance = widget.LowImportance
	return fullscreenButton
}

func (v *videoControlsRenderer) createPlayButton() *widget.Button {
	playbutton := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		if v.parent.viewer.pipeline == nil {
			return
		}
		if v.parent.viewer.IsPlaying() {
			v.parent.viewer.Pause()
		} else {
			v.parent.viewer.Play()
		}
		if v.parent.onTapped == nil {
			return
		}
		v.parent.onTapped()
	})
	playbutton.Importance = widget.LowImportance
	return playbutton
}

func (v *videoControlsRenderer) createStepBackwardButton() *widget.Button {
	stepBackwardButton := widget.NewButtonWithIcon("", theme.MediaFastRewindIcon(), func() {
		pos, _ := v.parent.viewer.CurrentPosition()
		v.parent.viewer.Seek(pos - v.parent.timeStep)
		if v.parent.onTapped == nil {
			return
		}
		v.parent.onTapped()
	})
	stepBackwardButton.Importance = widget.LowImportance
	return stepBackwardButton
}

func (v *videoControlsRenderer) createStepForwardButton() *widget.Button {
	stepForwardButton := widget.NewButtonWithIcon("", theme.MediaFastForwardIcon(), func() {
		pos, _ := v.parent.viewer.CurrentPosition()
		v.parent.viewer.Seek(pos + v.parent.timeStep)
		if v.parent.onTapped == nil {
			return
		}
		v.parent.onTapped()
	})
	stepForwardButton.Importance = widget.LowImportance
	return stepForwardButton
}

func (v *videoControlsRenderer) createVolumeMuteButton() *widget.Button {
	volumeMuteButton := widget.NewButtonWithIcon("", theme.VolumeUpIcon(), func() {
		v.parent.viewer.ToggleMute()
		if v.parent.viewer.IsMuted() {
			v.muteButton.SetIcon(theme.VolumeMuteIcon())
		} else {
			v.muteButton.SetIcon(theme.VolumeUpIcon())
		}
		if v.parent.onTapped == nil {
			return
		}
		v.parent.onTapped()
	})
	volumeMuteButton.Importance = widget.LowImportance
	return volumeMuteButton
}

func (v *videoControlsRenderer) createVolumeSlider() *widget.Slider {

	volumeSlider := widget.NewSlider(0, 1)
	volumeSlider.Orientation = widget.Vertical
	volumeSlider.Step = 0.01
	volumeSlider.Value = 1
	volumeSlider.OnChanged = func(value float64) {
		v.parent.viewer.SetVolume(value)
	}

	return volumeSlider
}

// showVideoControls displays the video controls dialog to control the video balance (contrast, brightness, hue and saturation).
func (v *videoControlsRenderer) showVideoControls() {
	if v.parent.viewer.pipeline == nil {
		return
	}

	// Some Labels to display the current values
	cl := widget.NewLabel("Contrast\n0.00")
	bl := widget.NewLabel("Brightness\n0.00")
	hl := widget.NewLabel("Hue\n0.00")
	sl := widget.NewLabel("Saturation\n0.00")
	for _, l := range []*widget.Label{cl, bl, hl, sl} {
		l.Alignment = fyne.TextAlignCenter
	}

	// The sliders to control the values
	hueSlider := widget.NewSlider(-1, 1)
	hueSlider.Value = 0
	hueSlider.Step = 0.01
	hueSlider.Orientation = widget.Vertical
	hueSlider.OnChanged = func(value float64) {
		v.parent.viewer.SetHue(value)
		hl.SetText(fmt.Sprintf("Hue\n%.2f", value))
	}

	contrastSlider := widget.NewSlider(0, 2)
	contrastSlider.Value = 1
	contrastSlider.Step = 0.01
	contrastSlider.Orientation = widget.Vertical
	contrastSlider.OnChanged = func(value float64) {
		v.parent.viewer.SetContrast(value)
		cl.SetText(fmt.Sprintf("Contrast\n%.2f", value))
	}

	saturationSlider := widget.NewSlider(0, 2)
	saturationSlider.Value = 1
	saturationSlider.Step = 0.01
	saturationSlider.Orientation = widget.Vertical
	saturationSlider.OnChanged = func(value float64) {
		v.parent.viewer.SetSaturation(value)
		sl.SetText(fmt.Sprintf("Saturation\n%.2f", value))
	}

	brightnessSlider := widget.NewSlider(-1, 1)
	brightnessSlider.Value = 0
	brightnessSlider.Step = 0.01
	brightnessSlider.Orientation = widget.Vertical
	brightnessSlider.OnChanged = func(value float64) {
		v.parent.viewer.SetBrightness(value)
		bl.SetText(fmt.Sprintf("Brightness\n%.2f", value))
	}

	resetIcon := theme.ContentUndoIcon()
	creset := widget.NewButtonWithIcon("", resetIcon, func() {
		contrastSlider.SetValue(1)
	})
	breset := widget.NewButtonWithIcon("", resetIcon, func() {
		brightnessSlider.SetValue(0)
	})
	hreset := widget.NewButtonWithIcon("", resetIcon, func() {
		hueSlider.SetValue(0)
	})
	sreset := widget.NewButtonWithIcon("", resetIcon, func() {
		saturationSlider.SetValue(1)
	})
	for _, b := range []*widget.Button{creset, breset, hreset, sreset} {
		b.Importance = widget.LowImportance
	}

	// get the current window, managing the fullscreen mode
	currentWindow := v.parent.viewer.currentWindowFinder()

	// create a dialog to display the controls
	// TODO: the dialog is modal and so a background is displayed, it alters the video view
	controlDialog := dialog.NewCustom("Video Controls", "Close", container.NewBorder(
		container.NewGridWithColumns(4, cl, bl, hl, sl),                 // top
		container.NewGridWithColumns(4, creset, breset, hreset, sreset), // center
		nil, // left
		nil, // right
		container.NewGridWithColumns(4,
			contrastSlider, brightnessSlider, hueSlider, saturationSlider,
		),
	), currentWindow)

	controlDialog.Resize(fyne.NewSize(200, 400)) // TODO: arbitrary size

	// assgin current values
	contrast, brightness, hue, saturation := v.parent.viewer.GetVideoBalance()

	contrastSlider.SetValue(contrast)
	brightnessSlider.SetValue(brightness)
	hueSlider.SetValue(hue)
	saturationSlider.SetValue(saturation)

	contrastSlider.OnChanged(contrast)
	brightnessSlider.OnChanged(brightness)
	hueSlider.OnChanged(hue)
	saturationSlider.OnChanged(saturation)

	controlDialog.Show()
}
