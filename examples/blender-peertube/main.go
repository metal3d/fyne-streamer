package main

import (
	_ "image/jpeg"
	_ "image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/metal3d/fyne-streamer/video"
)

func main() {

	a := app.New()
	w := a.NewWindow("Simple video player made with Fyne.io")

	videoWidget := video.NewPlayer()

	// if the widget is able to get the title from metadata,
	// it's nice to set it as the window title.
	videoWidget.SetOnTitle(func(title string) {
		w.SetTitle(title)
	})

	// A button to select a file to play
	selectFileButton := widget.NewButton("Select file", func() {
		opendialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			if err != nil {
				return
			}
			if file == nil {
				return
			}
			w.SetTitle(file.URI().Name())
			videoWidget.Open(file.URI())
			videoWidget.Play()
		}, w)
		opendialog.SetFilter(storage.NewExtensionFileFilter([]string{".mp4", ".mkv", ".avi", ".webm"}))
		opendialog.Show()
	})

	// a URL entry field to let the user enter a video URL
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Enter URL of a video")
	urlEntry.OnSubmitted = func(s string) {
		u, err := storage.ParseURI(s)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if err := videoWidget.Open(u); err != nil {
			dialog.ShowError(err, w)
			return
		}
		w.SetTitle(u.Name())
		videoWidget.Play()
	}

	// title for the list of videos from PeerTube
	title := widget.NewLabel("Official Blender Open Movies - PeerTube")
	title.TextStyle.Bold = true
	title.Alignment = fyne.TextAlignCenter

	// downloading videos from PeerTube could be slow, so we put it in a go routine
	// and display a progress bar while it's loading
	go func() {
		hsplit := container.NewHSplit(container.NewBorder(
			title,
			nil,
			nil,
			nil,
			getPeerTubeList(videoWidget, w),
		), videoWidget)
		hsplit.SetOffset(0.3)

		w.SetContent(container.NewBorder(
			nil,
			container.NewVBox(
				selectFileButton,
				urlEntry,
			),
			nil,
			nil,
			hsplit,
		))
	}()

	// progress bar while loading videos from PeerTube
	progress := widget.NewProgressBarInfinite()
	loading := widget.NewLabel("Downloading videos informations from PeerTube...")
	loading.Alignment = fyne.TextAlignCenter
	loading.TextStyle.Bold = true
	w.SetContent(container.NewBorder(
		nil,
		progress,
		nil,
		nil,
		loading,
	))

	// allow the user to drop a file on the window to play it
	w.SetOnDropped(func(p fyne.Position, u []fyne.URI) {
		if len(u) == 0 {
			return
		}

		// TODO: why not using go routine fails?
		go func() {
			videoWidget.Pause()
			videoWidget.Open(u[0])
			videoWidget.Play()
		}()
	})

	w.Resize(fyne.NewSize(1180, 750))
	w.ShowAndRun()
}
