// widgets/pie_chart.go
package widgets

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// palette is a fixed set of chart colours shared by pie and bar charts.
var palette = []color.NRGBA{
	{R: 37, G: 99, B: 235, A: 255},  // blue-600
	{R: 234, G: 88, B: 12, A: 255},  // orange-600
	{R: 22, G: 163, B: 74, A: 255},  // green-600
	{R: 147, G: 51, B: 234, A: 255}, // purple-600
	{R: 220, G: 38, B: 38, A: 255},  // red-600
	{R: 202, G: 138, B: 4, A: 255},  // yellow-600
	{R: 20, G: 184, B: 166, A: 255}, // teal-500
	{R: 236, G: 72, B: 153, A: 255}, // pink-500
}

// PieChart renders a pie chart using canvas.Raster.
type PieChart struct {
	data   jira.ChartData
	raster *canvas.Raster
	canvas fyne.CanvasObject
}

// NewPieChart creates a pie chart for the given data.
func NewPieChart(data jira.ChartData) *PieChart {
	p := &PieChart{data: data}

	p.raster = canvas.NewRaster(func(w, h int) image.Image {
		return drawPie(p.data, w, h)
	})
	p.raster.SetMinSize(fyne.NewSize(300, 300))

	legend := container.NewVBox()
	total := 0.0
	for _, v := range data.Values {
		total += v
	}
	for i, label := range data.Labels {
		col := palette[i%len(palette)]
		dot := canvas.NewRectangle(col)
		dot.SetMinSize(fyne.NewSize(12, 12))
		pct := 0.0
		if total > 0 && i < len(data.Values) {
			pct = data.Values[i] / total * 100
		}
		legend.Add(container.NewHBox(dot, widget.NewLabel(fmt.Sprintf("%s (%.1f%%)", label, pct))))
	}

	p.canvas = container.NewHBox(p.raster, legend)
	return p
}

// Canvas returns the Fyne canvas object.
func (p *PieChart) Canvas() fyne.CanvasObject { return p.canvas }

// SetData updates the chart data and redraws.
func (p *PieChart) SetData(data jira.ChartData) {
	p.data = data
	p.raster.Refresh()
}

func drawPie(data jira.ChartData, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	total := 0.0
	for _, v := range data.Values {
		total += v
	}
	if total == 0 || len(data.Values) == 0 {
		return img
	}

	cx, cy := float64(w)/2, float64(h)/2
	r := math.Min(cx, cy) * 0.85
	startAngle := -math.Pi / 2

	for i, val := range data.Values {
		sweep := val / total * 2 * math.Pi
		col := palette[i%len(palette)]

		steps := int(sweep/(2*math.Pi)*360*2)
		if steps < 4 {
			steps = 4
		}
		for s := 0; s <= steps; s++ {
			angle := startAngle + sweep*float64(s)/float64(steps)
			x2 := cx + r*math.Cos(angle)
			y2 := cy + r*math.Sin(angle)
			drawLineRGBA(img, int(cx), int(cy), int(x2), int(y2), col)
		}
		startAngle += sweep
	}
	return img
}

// drawLineRGBA draws a line using Bresenham's algorithm.
func drawLineRGBA(img *image.RGBA, x0, y0, x1, y1 int, col color.NRGBA) {
	dx := absInt(x1 - x0)
	dy := absInt(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	for {
		img.Set(x0, y0, col)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
