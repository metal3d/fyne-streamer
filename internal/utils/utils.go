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
		initGstreamer()
		runGlibMainLoop()
	}
}

func initGstreamer() {
	gst.Init(nil)
	gstInit = true
}

func runGlibMainLoop() {
	go glib.NewMainLoop(glib.MainContextDefault(), false).Run()
}

// MustCreateElement creates a new element with the given name and properties. The properties can be nil.
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

// AddToPipeline is a convenience function to add all elements to the pipeline from a map built with the mustCreateElement function.
func AddToPipeline(pipeline *gst.Pipeline, elements map[string]*gst.Element) error {
	err := pipeline.AddMany(getElementsFromMap(elements)...)
	return err
}

func getElementsFromMap(elements map[string]*gst.Element) []*gst.Element {
	all := make([]*gst.Element, len(elements))
	i := 0
	for _, element := range elements {
		all[i] = element
		i++
	}
	return all
}

// RemoveComments removes all comments from the given string.
func RemoveComments(s string) string {
	s = removeHashComments(s)
	s = unindentAndRemoveEmpty(s)
	return s
}

func removeHashComments(s string) string {
	// remove all #...$ from the string, multiline
	s = regexp.MustCompile(`(?m)#.*$`).ReplaceAllString(s, "")
	return s
}

func unindentAndRemoveEmpty(s string) string {
	// unindent and remove empty
	s = regexp.MustCompile(`(?m)^\s*`).ReplaceAllString(s, "")
	return s
}
