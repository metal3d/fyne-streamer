package video

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/test"
	"github.com/metal3d/fyne-streamer/internal/utils"
	"github.com/stretchr/testify/assert"
)

func init() {
	utils.GstreamerInit()
}

func TestCreatePlayer(t *testing.T) {
	player := NewPlayer()
	_ = test.WidgetRenderer(player)
	pipeline := player.Pipeline()
	assert.Nil(t, pipeline)
}

func TestPlayerCursor(t *testing.T) {
	player := NewPlayer()
	window := test.NewWindow(player)
	window.Resize(fyne.NewSize(320, 240))
	window.Show()

	err := player.Open(storage.NewFileURI(_testVideoFile))
	assert.Nil(t, err)

	v := player.controls.renderer.cursor.Value
	assert.Equal(t, float64(0), v)

	player.Play()
	time.Sleep(1 * time.Second)

	v = player.controls.renderer.cursor.Value
	assert.True(t, v > 0)
}
