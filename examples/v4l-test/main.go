package main

import (
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/go-gst/go-gst/gst"
	streamer "github.com/metal3d/fyne-streamer"
	"github.com/metal3d/fyne-streamer/video"
)

func main() {
	a := app.New()
	w := a.NewWindow("Camera with Ripple Effect")

	var source string
	switch runtime.GOOS {
	case "linux":
		source = "v4l2src"
	case "darwin":
		source = "avfvideosrc"
	case "windows":
		source = "ksvideosrc"
	default:
		panic("unsupported platform")
	}

	pipeline := `
    # source element is v4l2src on linux, avfvideosrc on darwin and ksvideosrc on windows
    %s name=cam !
    videoconvert name=convert !
    videoscale !
    videoflip video-direction=horiz name=direction ! # flip video horizontally
    # <- this is where we will add the ripple effect
    jpegenc name={{ .ImageEncoderElementName }} !
    appsink name={{ .AppSinkElementName }}
    `
	viewer := video.NewViewer()
	err := viewer.SetPipelineFromString(fmt.Sprintf(pipeline, source))
	if err != nil {
		panic(err)
	}
	viewer.Play()

	rippleButton := widget.NewButton("Toggle Ripple", func() {
		toggleRippleFilter(viewer.Pipeline())
	})

	w.SetContent(container.NewBorder(nil, rippleButton, nil, nil, viewer))

	w.Resize(fyne.NewSize(640, 480))
	w.ShowAndRun()
}

// This is a simple example of how to add a filter to a pipeline.
// We will add a ripple effect to the video stream between the
// videoflip and jpegenc elements. Or remove it if it is already there.
func toggleRippleFilter(pipeline *gst.Pipeline) {
	if pipeline == nil {
		return
	}
	videoflip, _ := pipeline.GetElementByName("direction")
	encoder, _ := pipeline.GetElementByName(streamer.ImageEncoderElementName)
	ripple, err := pipeline.GetElementByName("ripples")

	if err != nil || ripple == nil {
		// create a ripple effect and place it between videoflip and jpegenc
		ripple, _ := gst.NewElementWithName("rippletv", "ripples")
		pipeline.Add(ripple)                // add ripple to pipeline
		videoflip.Unlink(encoder)           // disconnect videoflip from encoder
		videoflip.Link(ripple)              // connect videoflip to ripple
		ripple.Link(encoder)                // and ripple to encoder
		pipeline.SetState(gst.StatePlaying) // let's go
	} else {
		// ripple already exists, remove it
		ripple.SetState(gst.StateNull)      // stop ripple before remove it
		pipeline.Remove(ripple)             // remove ripple from pipeline
		videoflip.Link(encoder)             // connect videoscale to encoder
		pipeline.SetState(gst.StatePlaying) // let's go
	}
}
