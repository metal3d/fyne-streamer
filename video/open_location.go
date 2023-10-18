package video

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
)

// Open opens the given location. It can be a file URI, an http or https URL.
func (v *Viewer) Open(u fyne.URI) error {
	switch u.Scheme() {
	case "http", "https":
		return v.openURL(u)
	case "file":
		return v.openFile(u)
	default:
		return fmt.Errorf("unsupported scheme %q", u.Scheme())
	}
}

// OpenURL opens the given stream from http or https URL.
// The pipeline has this structure:
//
//	        +--------------+
//	        |  souphttpsrc  |
//	        +------+-------+
//	               ↓
//	        +------|-------+
//	        |  decodebin   |
//	        +------+-------+
//	           ↓         ↓
//	+--------------+   +--------------+
//	| videoconvert |   | audioconvert |
//	+------+-------+   +----+---------+
//	       ↓                ↓
//	+--------------+  +---------------+
//	|  videorate   |  | audioresample |
//	+------+-------+  +-----+---------+
//	       ↓                ↓
//	+-------------+   +---------------+
//	|  videoscale |   | autoaudiosink |
//	+------+------+   +---------------+
//	       ↓
//	+-------------+
//	|   jpegenc   |
//	+------+------+
//	       ↓
//	+-------------+
//	|   appsink   |
//	+-------------+
func (v *Viewer) openURL(location fyne.URI) error {
	pipeline := `
    # the video source is sent to decoder
    souphttpsrc name={{ .InputElementName }} location=%[1]q ! 
    decodebin name={{ .DecodeElementName }} use-buffering=true

    # manage the video
    {{ .DecodeElementName }}. !
    queue !
    videoconvert ! 
    videoscale !
    videorate name={{ .VideoRateElementName }} ! 
    videobalance name={{ .VideoBalanceElementName }} !
    jpegenc name={{ .ImageEncoderElementName }} ! 
    appsink name={{ .AppSinkElementName }} sync=true max-lateness=%[2]d

    # manage the sound
    {{ .DecodeElementName }}. ! 
    queue !
    audioconvert ! 
    audioresample ! 
    volume name={{ .VolumeElementName }}  !
    autoaudiosink sync=true
    `

	pipeline = fmt.Sprintf(
		pipeline,
		location.String(),
		int64(time.Second.Nanoseconds()/int64(v.rate)),
	)
	return v.SetPipelineFromString(pipeline)
}

// Open a video from file. The pipeline has this structure:
//
//	        +--------------+
//	        |   filsesrc   |
//	        +------+-------+
//	               ↓
//	        +------|-------+
//	        |  decodebin   |
//	        +------+-------+
//	               ↓
//	         +-----+-----+
//	         ↓           ↓
//	+--------------+   +--------------+
//	| videoconvert |   | audioconvert |
//	+------+-------+   +----+---------+
//	       ↓                ↓
//	+--------------+  +---------------+
//	|  videorate   |  | audioresample |
//	+------+-------+  +-----+---------+
//	       ↓                ↓
//	+-------------+   +---------------+
//	|   jpegenc   |   | autoaudiosink |
//	+------+------+   +---------------+
//	       ↓
//	+-------------+
//	|   appsink   |
//	+-------------+
//
// The appsink element provides the frames, and the audio is
// connected to the default audio output of the system.
func (v *Viewer) openFile(location fyne.URI) error {
	v.reset()

	pipeline := `
    # the video source is sent to decoder
    filesrc name={{ .InputElementName }} location=%[1]q !
    decodebin name={{ .DecodeElementName }} use-buffering=true

    # manage the video
    {{ .DecodeElementName }}. !
    queue max-size-buffers=0 max-size-time=%[2]d !
    videoconvert !
    videorate name={{ .VideoRateElementName }} !
    videoscale !
    videobalance name={{ .VideoBalanceElementName }} !
    jpegenc name={{ .ImageEncoderElementName }} !
    appsink name={{ .AppSinkElementName }} sync=true max-lateness=%[2]d

    # manage the sound
    {{ .DecodeElementName }}. !
    queue max-size-buffers=0 max-size-time=%[2]d !
    audioconvert !
    audioresample !
    volume name={{ .VolumeElementName }}  !
    autoaudiosink sync=true
    `
	pipeline = fmt.Sprintf(
		pipeline,
		location.Path(),
		int64(time.Second.Nanoseconds()/int64(v.rate)),
	)
	return v.SetPipelineFromString(pipeline)
}
