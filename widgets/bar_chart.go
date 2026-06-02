// widgets/bar_chart.go
package widgets

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// BarChart renders a horizontal bar chart using canvas.Raster.
type BarChart struct {
	data   jira.ChartData
	raster *canvas.Raster
	canvas fyne.CanvasObject
}

// NewBarChart creates a bar chart for the given data.
func NewBarChart(data jira.ChartData) *BarChart {
	b := &BarChart{data: data}

	b.raster = canvas.NewRaster(func(w, h int) image.Image {
		return drawBars(b.data, w, h)
	})
	b.raster.SetMinSize(fyne.NewSize(400, 250))

	labels := container.NewVBox()
	for i, label := range data.Labels {
		if i >= len(data.Values) {
			break
		}
		labels.Add(widget.NewLabel(fmt.Sprintf("%s: %.1fh", label, data.Values[i])))
	}

	b.canvas = container.NewVBox(b.raster, labels)
	return b
}

// Canvas returns the Fyne canvas object.
func (b *BarChart) Canvas() fyne.CanvasObject { return b.canvas }

// SetData updates the chart data and redraws.
func (b *BarChart) SetData(data jira.ChartData) {
	b.data = data
	b.raster.Refresh()
}

func drawBars(data jira.ChartData, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	n := len(data.Values)
	if n == 0 {
		return img
	}

	maxVal := 0.0
	for _, v := range data.Values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		return img
	}

	barHeight := h / (n + 1)
	if barHeight < 4 {
		barHeight = 4
	}
	padding := barHeight / 4

	for i, val := range data.Values {
		col := palette[i%len(palette)]
		barW := int(val / maxVal * float64(w-4))
		y0 := i*(barHeight+padding) + padding
		y1 := y0 + barHeight
		for y := y0; y < y1 && y < h; y++ {
			for x := 2; x < barW+2 && x < w; x++ {
				img.Set(x, y, col)
			}
		}
	}
	return img
}
