package video

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	streamer "github.com/metal3d/fyne-streamer"
	"github.com/metal3d/fyne-streamer/internal/utils"
)

// prerollFunc is called when the pipeline is prerolling. This is a callback on the appsink.
func (v *Viewer) prerollFunc(appSink *app.Sink) gst.FlowReturn {
	caps, err := appSink.Element.GetPads()
	if len(caps) == 0 && err != nil {
		return gst.FlowError
	}

	c := caps[0].GetCurrentCaps()
	ww, err := c.GetStructureAt(0).GetValue("width")
	if err != nil {
		log.Printf("Error: %v", err)
	}

	hh, err := c.GetStructureAt(0).GetValue("height")
	if err != nil {
		log.Printf("Error: %v", err)
	}

	if ww, ok := ww.(int); ok {
		v.width = ww
	}

	if hh, ok := hh.(int); ok {
		v.height = hh
	}

	// call the callback
	if v.onPreRoll != nil {
		v.onPreRoll()
	}
	return gst.FlowOK
}

func (v *Viewer) reset() {
	if v.pipeline != nil {
		//BUG: do not use v.SetState(gst.StateNull)
		if err := v.pipeline.SetState(gst.StateNull); err != nil {
			fyne.LogError("Failed to set pipeline to null", err)
		} else {
			v.pipeline.Clear()
		}
	}
	v.frame.Image = nil
	v.duration = 0
}

func (v *Viewer) registerElements() error {
	v.appSink = nil

	var err error
	_, err = v.pipeline.GetElementByName(streamer.InputElementName)
	if err != nil {
		log.Printf("Warning: Failed to find "+streamer.InputElementName+" element: %v", err)
	}

	_, err = v.pipeline.GetElementByName(streamer.VideoRateElementName)
	if err != nil {
		log.Printf("Warning: Failed to find "+streamer.VideoRateElementName+" element: %v", err)
	}

	_, err = v.pipeline.GetElementByName(streamer.ImageEncoderElementName)
	if err != nil {
		log.Printf("Warning: Failed to find "+streamer.ImageEncoderElementName+" element: %v", err)
	}

	_, err = v.pipeline.GetElementByName(streamer.DecodeElementName)
	if err != nil {
		log.Printf("Warning: Failed to find "+streamer.DecodeElementName+" element: %v", err)
	}

	appelement, err := v.pipeline.GetElementByName(streamer.AppSinkElementName)
	if err != nil || appelement == nil {
		fyne.LogError("Failed to find the appsink element", err)
		return fmt.Errorf("Failed to find the mandatory %s element %w", streamer.AppSinkElementName, err)
	}
	v.appSink = app.SinkFromElement(appelement)
	v.appSink.SetCallbacks(&app.SinkCallbacks{
		EOSFunc:        v.eosFunc,
		NewPrerollFunc: v.prerollFunc,
		NewSampleFunc:  v.newSampleFunc,
	})

	return nil
}

// eosFunc is called when the pipeline is at the end of the stream. This is a callback on the appsink.
func (v *Viewer) eosFunc(appSink *app.Sink) {
	if v.onEOS != nil {
		v.onEOS()
	}
	err := v.SetState(gst.StatePaused)
	if err != nil {
		fyne.LogError("Failed to set pipeline to paused", err)
	}
	// TODO: this is a workaround to avoid a crash when the pipeline is stopped and to make it restartable.
	time.AfterFunc(time.Millisecond*100, func() {
		v.Seek(0)
		time.AfterFunc(time.Millisecond*100, func() {
			img, err := v.getCurrentFrame(v.appSink, true)
			if err != gst.FlowOK {
				log.Println("error getting first frame", err)
				return
			}
			v.frame.Image = img
			v.frame.Refresh()
		})
	})
}

// newSampleFunc is called when a new sample is available. This is a callback on the appsink.
// The sample is a jpeg image and is decoded to the internal frameView canvas.Image.
// It is refreshed as soon as the frame is OK.
func (v *Viewer) newSampleFunc(appSink *app.Sink) gst.FlowReturn {

	img, ret := v.getCurrentFrame(appSink, false)
	if ret != gst.FlowOK {
		return ret
	}

	func() { // refresh the frame in a goroutine
		v.frame.Image = img
		v.frame.Refresh()
	}()

	_, pos := v.pipeline.QueryPosition(gst.FormatTime)
	if v.onNewFrame != nil {
		// get the current time of the pipeline
		go v.onNewFrame(time.Duration(float64(pos)))
	}

	return ret
}

func (v *Viewer) getCurrentFrame(appSink *app.Sink, latest bool) (image.Image, gst.FlowReturn) {
	var sample *gst.Sample
	if !latest {
		sample = appSink.PullSample()
	} else {
		sample = appSink.GetLastSample()
	}
	if sample == nil {
		return nil, gst.FlowEOS
	}

	buffer := sample.GetBuffer() // Get the buffer from the sample
	if buffer == nil {
		return nil, gst.FlowError
	}
	defer buffer.Unmap()

	samples := buffer.Map(gst.MapRead).AsUint8Slice()
	if samples == nil {
		return nil, gst.FlowError
	}

	// the sample is a jpeg
	reader := bytes.NewReader(samples)
	img, _, err := image.Decode(reader)
	if err != nil {
		fyne.LogError("Failed to decode image: %w", err)
		return nil, gst.FlowError
	}

	return img, gst.FlowOK
}

// resync the pipeline with the parent state. This seems to fix some problem
// on leaveing paused state. But not always...
// At this time, this function is called on SetState and Seek methods.
func (v *Viewer) resync() {
	// resync the pipeline
	if v.pipeline == nil {
		return
	}
	go func() {
		elements, _ := v.pipeline.GetElements()
		for _, e := range elements {
			done := e.SyncStateWithParent()
			if !done {
				log.Printf("Warning: Failed to sync state with parent for element %s", e.GetName())
			}
		}
	}()
}

// SetState sets the state of the pipeline. It is a blocking call with a timeout of 50ms. This is a workaround
// because the pipeline.SetState can block forever if the pipeline is not in a good state.
func (v *Viewer) setState(state gst.State) error {

	if v.pipeline == nil {
		return fmt.Errorf("No pipeline")
	}

	// change the state
	err := v.pipeline.SetState(state)
	if err != nil {
		fyne.LogError("Failed to set state", err)
		return err
	}

	// wait for the state to be set
	done := make(chan error)
	go func(done chan error) {
		var err error
		defer func() {
			done <- err
		}()
		for {
			select {
			case <-time.Tick(1000 * time.Millisecond):
				err = fmt.Errorf("timeout waiting for state %s", state)
				return
			case <-time.After(1 * time.Millisecond):
				current_state := v.pipeline.GetCurrentState()
				if current_state == state {
					return
				}
			}
		}
	}(done)

	return <-done
}

func (v *Viewer) setCurrentWindowFinder(w fyne.CanvasObject) {
	v.currentWindowFinder = func() fyne.Window {
		return utils.GetWindowForElement(w)
	}
}

// fullscreenOff sets the video widget to windowed mode.
func (v *Viewer) fullscreenOff() {
	if v.fullscreenWindow == nil {
		fyne.LogError("Could not find fullscreen window", nil)
		return
	}
	v.fullscreenWindow.Close()
	if v.currentWindow == nil {
		fyne.LogError("Could not find current window", nil)
		return
	}
}

// fullscreenOn sets the video widget to fullscreen.
func (v *Viewer) fullscreenOn() {
	if v.currentWindow == nil {
		v.currentWindow = v.currentWindowFinder()
		if v.currentWindow == nil {
			fyne.LogError("Could not find current window", nil)
			return
		}
	}

	fsWindow := fyne.CurrentApp().NewWindow("")
	v.fullscreenWindow = fsWindow

	// ESC or F key to exit fullscreen
	fsWindow.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		if k.Name == fyne.KeyEscape || k.Name == fyne.KeyF {
			v.fullscreenOff()
		}
	})

	// set the content of the fullscreen window to the video widget
	if v.originalViewerWidget == nil {
		v.originalViewerWidget = v
	}

	// make a black rectangle to fill the background
	background := canvas.NewRectangle(color.Black)

	// and set the content of the fullscreen window to the video widget + black background
	fsWindow.SetContent(
		container.NewStack(
			background,
			v.originalViewerWidget,
		),
	)

	// on close, we show the app window and refresh the content
	fsWindow.SetOnClosed(func() {
		v.currentWindow.Show()
		v.currentWindow.Content().Refresh()
		// BUG: this is a workaround to force the cursor to go back to the right position. It make a little flicker, but it works...
		if pos, err := v.CurrentPosition(); err == nil {
			v.Seek(pos)
		}
	})

	// hide the current window and show the fullscreen window
	v.currentWindow.Hide()

	// make the new window fullscreen, unpadded and show it
	fsWindow.SetFullScreen(true)
	fsWindow.SetPadded(false)
	fsWindow.Show()
}
