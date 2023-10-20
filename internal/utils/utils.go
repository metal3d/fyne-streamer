package utils

import (
	"regexp"

	"fyne.io/fyne/v2"
	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
)

var (
	gstInit = false
)

// GetWindowForElement returns the window that contains the given element.
func GetWindowForElement(o fyne.CanvasObject) fyne.Window {
	canvas := fyne.CurrentApp().Driver().CanvasForObject(o)
	windows := fyne.CurrentApp().Driver().AllWindows()
	for _, w := range windows {
		if w.Canvas() == canvas {
			return w
		}
	}
	return nil
}

// GstreamerInit initializes gstreamer if it is not already initialized.
func GstreamerInit() {
	if !gstInit {
		gst.Init(nil)
		gstInit = true
		go glib.NewMainLoop(glib.MainContextDefault(), false).Run() //
	}
}

// MustCreateElement creates a new element with the given name and properties. The propoerties can be nil.
func MustCreateElement(elementname string, properties map[string]interface{}) *gst.Element {
	if properties == nil {
		element, err := gst.NewElement(elementname)
		if err != nil {
			panic(err)
		}
		return element
	} else {
		element, err := gst.NewElementWithProperties(elementname, properties)
		if err != nil {
			panic(err)
		}
		return element
	}
}

// NewPipelineFromString creates a new pipeline from the given string.
func NewPipelineFromString(pipeline string) (*gst.Pipeline, error) {
	GstreamerInit()
	return gst.NewPipelineFromString(RemoveComments(pipeline))
}

// convenience function to add all elements to the pipeline from a map
// built with the mustCreateElement function.
func AddToPipeline(pipeline *gst.Pipeline, elements map[string]*gst.Element) error {
	err := pipeline.AddMany(func() []*gst.Element {
		all := make([]*gst.Element, len(elements))
		i := 0
		for _, element := range elements {
			all[i] = element
			i++
		}
		return all
	}()...)
	return err
}

// RemoveComments removes all comments from the given string.
func RemoveComments(s string) string {
	// remove all #...$ from the string, multiline
	s = regexp.MustCompile(`(?m)#.*$`).ReplaceAllString(s, "")
	// unindent and remove empty
	s = regexp.MustCompile(`(?m)^\s*`).ReplaceAllString(s, "")
	return s
}
