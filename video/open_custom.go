package video

import (
	"bytes"
	"text/template"

	"github.com/go-gst/go-gst/gst"
	streamer "github.com/metal3d/fyne-streamer"
	"github.com/metal3d/fyne-streamer/internal/utils"
)

// SetPipelineFromString creates a pipeline from a string. The string is a GStreamer pipeline description
// that is a template (text/template) with the ElementMap as data. So you can use the
// ElementName constants in the template.
//
// It also upports comments starting with #.
//
// It is mandatory to provide at least an "appsink" element named with AppSinkElementName.
//
// You can also provide other elements that the viewer will manage.
// Use provided const names for the elements: InputElementName, DecodeElementName,
// VideoRateElementName, ImageEncoderElementName, AppSinkElementName.
//
// For example, getting the webcam stream can be done with:
//
//	pipeline := `
//	    autovideosrc name={{ .InputElementName }} ! # the input, a video test
//	    videoconvert n-threads=4 ! # convert to something usable
//	    videorate name={{ .VideoRateElementName }} max-rate=30 ! # fix the framerate
//	    # encode to jpeg (or png), mandatory for appsink
//	    jpegenc name={{ .ImageEncoderElementName }} quality=80 !
//	    # the appsink (mandatory)
//	    appsink name={{ .AppSinkElementName }} drop=true max-lateness=33333 sync=true
//	    `
//	v.Custom(pipeline)
//
// Note that the appsink element max-lateness property is set to 33333 nanoseconds, which is the equivalent of 30 fps. This argument is modified by the Video element if needed.
func (v *Viewer) SetPipelineFromString(pipeline string) error {

	v.reset()

	tpl, err := template.New("pipeline").Parse(pipeline)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, streamer.ElementMap)
	if err != nil {
		return err
	}

	pipeline = buf.String()
	pipelineObj, err := utils.NewPipelineFromString(pipeline)
	if err != nil {
		return err
	}

	v.pipeline = pipelineObj

	v.createBus()
	return v.registerElements()
}

// SetPipeline creates a pipeline from a gst.Pipeline object. As for CustomFromString,
// it is mandatory to provide at least an "appsink" element named with AppSinkElementName.
//
// You can also provide other elements that the viewer will manage.
// Use provided const names for the elements: InputElementName, DecodeElementName,
// VideoRateElementName, ImageEncoderElementName, AppSinkElementName.
func (v *Viewer) SetPipeline(p *gst.Pipeline) error {
	v.reset()
	v.pipeline = p
	v.createBus()
	return v.registerElements()
}

func (v *Viewer) createBus() {
	if v.pipeline == nil {
		return
	}
	bus := v.pipeline.GetPipelineBus()
	bus.AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageTag:
			tags := msg.ParseTags()
			title, ok := tags.GetString(gst.TagTitle)
			if v.onTitle != nil && ok {
				v.onTitle(title)
			}
		}
		return true
	})
	v.bus = bus
}
