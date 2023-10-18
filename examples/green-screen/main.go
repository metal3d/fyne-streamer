package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/metal3d/fyne-streamer/video"
)

const (
	jcvd = "https://peertube.fr/download/videos/25f555e2-7c9c-4504-9c3d-bf29d61dc836-360.mp4"
	doc  = `
## Green screen removal

We present here 2 players. The original video is opened with a simple Viewer.
The second one is a Viewer with a custom pipeline that removes the green screen and 
no sound to avoid to duplicate the audio.

The pipeline that removes the green screen is the following (except the parameters):

    souphttpsrc → decodebin → videoconvert → videorate → alpha → pngenc → appsink

The result is not perfect, the original video is a bit pixelated, and we don't tried to 
improve the result. But it works. And it works in real time.

Please refer to the source code to see the parameters
`
)

func main() {
	pipeline := `
    souphttpsrc location=%q name={{ .InputElementName }} ! 
    decodebin name={{ .DecodeElementName }} !
    videoconvert ! 
    videorate name={{ .VideoRateElementName }} max-rate=30 !
    
    # a color picker gave the background color as rgb(12,158,37)
    alpha
        method=custom # we will use a custom method to set our own background color
        angle=85      # because... don't ask me, it works...
        target-r=12 
        target-g=158 
        target-b=37 !

    # see, we are using alpha channel here, so we need to convert to PNG
    pngenc name={{ .ImageEncoderElementName }} ! 

    # the appsink (mandatory)
    appsink name={{ .AppSinkElementName }} sync=true

    # We don't connect the audio, the second player will play it.
    # This avoid to have 2 audio streams playing at the same time.
    `

	a := app.New()
	w := a.NewWindow("Video Player")

	greenScreenRemover := video.NewViewer()
	err := greenScreenRemover.SetPipelineFromString(fmt.Sprintf(pipeline, jcvd))
	if err != nil {
		panic(err)
	}

	originalVideo := video.NewViewer()
	uri, _ := storage.ParseURI(jcvd)
	originalVideo.Open(uri)

	playAll := func() {
		originalVideo.Seek(0)
		greenScreenRemover.Seek(0)
		originalVideo.Play()
		greenScreenRemover.Play()
	}

	playButton := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), playAll)
	playAll()

	originalLabel := widget.NewLabel("Original")
	greenScreenRemoverLabel := widget.NewLabel("Alpha set on green screen")
	for _, l := range []*widget.Label{originalLabel, greenScreenRemoverLabel} {
		l.Alignment = fyne.TextAlignCenter
	}

	doc := widget.NewRichTextFromMarkdown(doc)
	doc.Wrapping = fyne.TextWrapWord

	w.SetContent(
		container.NewBorder(doc, playButton, nil, nil,
			container.NewGridWithColumns(2,
				container.NewBorder(originalLabel, nil, nil, nil, originalVideo),
				container.NewBorder(greenScreenRemoverLabel, nil, nil, nil, greenScreenRemover),
			),
		),
	)

	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}
