package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/metal3d/fyne-streamer/video"
)

func main() {
	pipeline := `
    videotestsrc name={{ .InputElementName }} ! # the input, a video test
    videoconvert n-threads=4 ! # convert to something usable
    videorate name={{ .VideoRateElementName }} max-rate=30 ! # fix the framerate
    # encode to jpeg (or png), mandatory for appsink
    jpegenc name={{ .ImageEncoderElementName }} quality=80 !
    # the appsink (mandatory)
    appsink name={{ .AppSinkElementName }} drop=true max-lateness=33333 sync=true
    `

	a := app.New()
	w := a.NewWindow("Simple video test")

	videoWidget := video.NewViewer()
	videoWidget.SetPipelineFromString(pipeline)
	videoWidget.Play()

	w.Resize(fyne.NewSize(640, 480))
	w.SetContent(videoWidget)
	w.ShowAndRun()

}
