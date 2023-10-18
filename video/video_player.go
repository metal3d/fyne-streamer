package video

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// check that the Player implements the interfaces
var _ fyne.Widget = (*Player)(nil)
var _ desktop.Hoverable = (*Player)(nil)
var _ fyne.Tappable = (*Player)(nil)
var _ fyne.DoubleTappable = (*Player)(nil)

// Player is a Viewer widget with controls and interaction. This widget
// proposes auto-hidden controls, cursor to navigate in the video and,
// and many others features.
type Player struct {
	*Viewer
	autoHideTimer   time.Duration // time to wait before hiding the controls
	autoHide        bool
	cancelAutoHide  context.CancelFunc // cancel the autoHide goroutine
	autoHideContext context.Context    // context of the autoHide goroutine
	controls        *VideoControls     // controls of the video widget
}

// NewPlayer returns a new video widget with controls and interaction.
func NewPlayer() *Player {
	v := &Player{
		Viewer:        CreateBaseVideoViewer(),
		autoHideTimer: time.Second * 3,
		autoHide:      true,
	}
	v.ExtendBaseWidget(v)

	v.SetOnPreRoll(func() {
		duration, _ := v.Duration()
		v.controls.SetDuration(duration)
	})

	v.SetOnNewFrame(func(d time.Duration) {
		if v.controls == nil {
			return
		}
		v.controls.SetCursorAt(d)
		v.controls.onTapped = func() {
			if v.autoHide {
				v.controls.Show()
				v.doAutoHide()
			}
		}
	})

	if v.autoHide {
		v.doAutoHide()
	}

	return v
}

// CreateRenderer creates a renderer for the video widget.
//
// Implements: fyne.Widget
func (v *Player) CreateRenderer() fyne.WidgetRenderer {
	v.controls = NewVideoControls(v.Viewer)
	return widget.NewSimpleRenderer(
		container.NewStack(
			v.Frame(),
			v.controls,
		),
	)
}

// DoubleTapped toggles the video widget between playing and paused.
//
// Implements: fyne.DoubleTappable
func (v *Player) DoubleTapped(ev *fyne.PointEvent) {
	if v.Pipeline() == nil {
		return
	}
	// pause video or play
	if v.IsPlaying() {
		v.Pause()
	} else {
		v.Play()
	}
}

// MouseIn shows the controls of the video widget.
//
// Implements: desktop.Hoverable
func (v *Player) MouseIn(pos *desktop.MouseEvent) {}

// MouseMoved shows the controls of the video widget.
//
// Implements: desktop.Hoverable
func (v *Player) MouseMoved(pos *desktop.MouseEvent) {
	v.controls.Show()
	v.doAutoHide()
}

// MouseOut hides the controls of the video widget.
//
// Implements: desktop.Hoverable
func (v *Player) MouseOut() {}

// EnableAutoHide sets the autoHide feature of the video widget.
func (v *Player) EnableAutoHide(b bool) {
	v.autoHide = b
	v.controls.Show()
	if b {
		v.doAutoHide()
	} else {
		v.cancelAutoHide()
	}
}

// SetAutoHideTimer sets the time to wait before hiding the controls.
func (v *Player) SetAutoHideTimer(d time.Duration) {
	v.autoHideTimer = d
}

// Tapped hides the controls of the video widget.
//
// Implements: fyne.Tappable
func (v *Player) Tapped(ev *fyne.PointEvent) {
	if v.controls == nil {
		return
	}
	if !v.autoHide {
		return
	}
	if v.controls.Visible() {
		v.controls.Hide()
	} else {
		v.controls.Show()
		v.doAutoHide()
	}
}

// createContext creates a cancelable context for the autoHide goroutine.
// It cancels the previous context if it exists.
func (v *Player) createContext() {
	if v.cancelAutoHide != nil {
		v.cancelAutoHide()
	}
	v.autoHideContext, v.cancelAutoHide = context.WithCancel(
		context.Background(),
	)
}

// doAutoHide hides the controls after the autoHideTimer duration.
// This function cancels and creates a new cancelable context each time it is called (using Player.createContext).
func (v *Player) doAutoHide() {
	if !v.autoHide {
		return
	}
	if v.autoHideTimer <= 0 {
		return
	}
	v.createContext()
	go func() {
		select {
		case <-v.autoHideContext.Done(): // canceled, so do not hide
			return
		case <-time.After(v.autoHideTimer):
			v.controls.Hide()
		}
	}()
}
