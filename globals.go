package streamer

const (
	// InputElementName is the name of the input element. Filesrc, souphttpsrc...
	InputElementName ElementName = "fyne-input"

	// DecodeElementName is the name of the decode element. Actually, a decodebin.
	DecodeElementName ElementName = "fyne-decode"

	// AppSinkElementName is the name of the appsink element. It's the mandatory element
	// in the pipeline.
	AppSinkElementName ElementName = "fyne-app"

	// VideoRateElementName is the name of the videorate element. It's used to limit
	// the framerate of the video. Place it just after the videoconvert element in the same
	// branch of the appsink element
	VideoRateElementName ElementName = "fyne-videorate"

	// ImageEncoderElementName is the name of the image encoder element. It's used to
	// encode the video frames to jpeg or png. It must be the last element in the pipeline before
	// the appsink element.
	// Generally, it's a "jpenenc". But it can be a "pngenc" if you want to encode the frames
	// with alpha channel.
	ImageEncoderElementName ElementName = "fyne-imageencoder"

	// VolumeElementName is the name of the volume element.
	// It's used to control the volume of the audio.
	VolumeElementName ElementName = "fyne-volume"

	// VideoBalanceElementName is the name of the videobalance element. It can be used to
	// controle the brightness, contrast, hue and saturation of the video.
	VideoBalanceElementName ElementName = "fyne-videobalance"
)

// TimeFormat is the format used to display the time in the video widget.
const TimeFormat = "15:04:05"

// ElementMap is a map of the element names used in the pipeline. This is used
// in templates to create the pipeline.
var ElementMap = map[string]ElementName{
	"InputElementName":        InputElementName,
	"DecodeElementName":       DecodeElementName,
	"VideoRateElementName":    VideoRateElementName,
	"ImageEncoderElementName": ImageEncoderElementName,
	"AppSinkElementName":      AppSinkElementName,
	"VolumeElementName":       VolumeElementName,
	"VideoBalanceElementName": VideoBalanceElementName,
}

// ElementName is the name of a GStreamer element. It's a string (alias).
type ElementName = string
