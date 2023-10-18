/*
Package video proposes widgets to play video files using GStreamer. There are

two widgets: Player and Viewer. The Player widget is a complete widget that
can be used as is. The Viewer widget is a lower level widget that can be used
to create a custom video player or a simple video viewer with no graphical controls.

If you need a simple video viwer, use the Viewer widget.

	video := video.NewViewer()
	video.Open(uri)

If you need a video player with controls, use the Player widget.

	video := video.NewPlayer()
	video.Open(uri)
	video.Play()

Both widgets can be fullscreened and have Play(), Pause() and Seek(duration) methods.
The difference is that the Player has controls (auto hidden) and react on tap and double tap.

You can create your own Gstreamer pipeline and use the Viewer widget to display the video frames.
The mandatory element to create is an "appsink" that is name with "AppSinkElementName" (constant).
Others names can be provided to let the player adapt the framerate, the quality of the
image compression, etc. The encoders can be "jpegenc" or "pngenc" and it's mandatory to
have one of them in the pipeline before the appsink element.

To ease the creation of the pipeline, the CustomFromString method accepts a template string with
comments. The ElementMap variable is used as data for the template. So you can use the
ElementName constants in the template.

For example:

	pipeline := `
	# the video source is sent to decoder
	videotestsrc name={{ .InputElementName }} !
	videoconvert !
	videorate name={{ .VideoRateElementName }} !

	# add an image encoder here
	jpegenc name={{ .ImageEncoderElementName }} !

	# the appsink element is mandatory
	appsink name={{ .AppSinkElementName }} sync=true
	`
	video := video.NewViewer()
	video.CustomFromString(pipeline)
	video.Play()
*/
package video
