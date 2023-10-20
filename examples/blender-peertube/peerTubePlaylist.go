package main

import (
	"encoding/json"
	"fmt"
	"image"
	"log"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/metal3d/fyne-streamer/video"
)

const peerTubeURL = "https://video.blender.org"

var _ fyne.CanvasObject = (*videoListElement)(nil)
var _ fyne.Widget = (*videoListElement)(nil)

type peerTubeVideo struct {
	Name          string `json:"name"`
	Id            uint   `json:"id"`
	ThumbnailPath string `json:"thumbnailPath"`
	Thumbnail     image.Image
	VideoURL      string
}
type videoListElement struct {
	widget.BaseWidget
	Label *widget.Label
	Image *canvas.Image
}

func newVideoListElement() *videoListElement {
	v := &videoListElement{
		Label: widget.NewLabel(""),
		Image: canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1))),
	}
	v.ExtendBaseWidget(v)
	v.Label.Wrapping = fyne.TextWrapWord
	v.Image.FillMode = canvas.ImageFillContain
	v.Image.ScaleMode = canvas.ImageScaleFastest
	v.Image.SetMinSize(fyne.NewSize(90, 75))
	return v
}

func (v *videoListElement) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewBorder(
			nil, nil,
			v.Image, nil,
			v.Label,
		),
	)
}

func (v *videoListElement) SetVideoInfo(video *peerTubeVideo) {
	v.Label.SetText(video.Name)
	v.Image.Image = video.Thumbnail
}

func getPeerTubeList(videoWidget *video.Player, w fyne.Window) *widget.List {
	// get some videos from Blender PeerTube instance
	videos := getPeerTubeVideos()
	videotitles := func() []string {
		titles := make([]string, len(videos))
		i := 0
		for k := range videos {
			titles[i] = k
			i++
		}
		return titles
	}()

	list := widget.NewList(func() int {
		return len(videos)
	}, func() fyne.CanvasObject {
		return newVideoListElement()
	}, func(id widget.ListItemID, element fyne.CanvasObject) {
		e := element.(*videoListElement)
		video := videos[videotitles[id]]
		e.SetVideoInfo(&video)
	})
	list.OnSelected = func(id widget.ListItemID) {
		videoWidget.Pause()
		u, err := storage.ParseURI(videos[videotitles[id]].VideoURL)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		w.SetTitle(videotitles[id])
		videoWidget.Open(u)
		videoWidget.Play()
	}

	return list
}

func getPeerTubeVideos() map[string]peerTubeVideo {
	resp, err := http.Get(peerTubeURL + "/api/v1/video-channels/blender_open_movies/videos?count=100&sort=publishedAt")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	videos := struct {
		Data []peerTubeVideo `json:"data"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&videos)
	if err != nil {
		log.Fatal(err)
	}
	videoMap := make(map[string]peerTubeVideo)
	for _, video := range videos.Data {
		video.VideoURL = getVideoUrl(video.Id)
		resp, err := http.Get(peerTubeURL + video.ThumbnailPath)
		if err == nil {
			defer resp.Body.Close()
			video.Thumbnail, _, err = image.Decode(resp.Body)
			if err != nil {
				fyne.LogError("Failed to decode thumbnail", err)
			}
		}
		videoMap[video.Name] = video
	}
	return videoMap
}

func getVideoUrl(id uint) string {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/videos/%d", peerTubeURL, id))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	video := struct {
		Files []struct {
			FileUrl string `json:"fileUrl"`
		} `json:"files"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&video)
	if err != nil {
		log.Fatal(err)
	}
	if len(video.Files) > 0 {
		return video.Files[0].FileUrl
	}
	return ""
}
