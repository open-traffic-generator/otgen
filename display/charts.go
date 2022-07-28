package display

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"

	"github.com/lucasb-eyer/go-colorful"
)

type ChartProcessor struct {
	cnt    *container.Container
	series map[string]*ChartSeries
}

type ChartSeries struct {
	Data      []float64
	Name      string
	Chart     *linechart.LineChart
	MaxPoints int
	Color     cell.Color
}

func NewSeries(name string, maxpoints int, chart *linechart.LineChart, color cell.Color) *ChartSeries {
	return &ChartSeries{
		Data:      []float64{},
		Name:      name,
		Chart:     chart,
		MaxPoints: maxpoints,
		Color:     color,
	}
}

func (c *ChartSeries) AddPoint(val float64) error {
	if len(c.Data) == c.MaxPoints {
		c.Data = c.Data[1:]
	}
	c.Data = append(c.Data, val)
	return c.Chart.Series(c.Name, c.Data, linechart.SeriesCellOpts(cell.FgColor(c.Color)))
}

func (cp *ChartProcessor) Process(data []DataPoint) error {
	for _, p := range data {
		for k, v := range p {
			if k != NAME_FIELD {
				series_key := fmt.Sprintf("%s.%s", k, p[NAME_FIELD])
				var err error
				switch val := v.(type) {
				case float64:
					err = cp.series[series_key].AddPoint(val)
				case int:
					err = cp.series[series_key].AddPoint(float64(val))
				case string:
					value, perr := strconv.ParseFloat(val, 64)
					if perr != nil {
						return perr
					}
					err = cp.series[series_key].AddPoint(value)
				default:
					err = cp.series[series_key].AddPoint(math.NaN())
				}
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (cp *ChartProcessor) Layout(data []DataPoint) error {
	// Looks ugly, but in 10 minutes a better idea didn't come up
	// So, I shall live with this abomination of the code for the moment.
	cp.series = map[string]*ChartSeries{}
	charts := map[string]*ChartSeries{}
	for _, p := range data {
		for k := range p {
			if k != NAME_FIELD {
				series_key := fmt.Sprintf("%s.%s", k, p[NAME_FIELD])
				if _, ok := cp.series[series_key]; !ok {
					if _, ok := charts[k]; !ok {
						ch, _ := linechart.New()
						charts[k] = NewSeries(k, 100, ch, cell.ColorDefault)
					}
					color := colorful.HappyColor()

					cp.series[series_key] = &ChartSeries{
						Data:      []float64{},
						Name:      p[NAME_FIELD].(string),
						Chart:     charts[k].Chart,
						MaxPoints: charts[k].MaxPoints,
						Color:     cell.ColorRGB24(int(255.0*color.R), int(255.0*color.G), int(255.0*color.B)),
					}
				}
			}
		}
	}

	builder := grid.New()
	for k, v := range charts {
		builder.Add(grid.RowHeightPerc(99/len(charts), grid.Widget(v.Chart, container.Border(linestyle.Light), container.BorderTitle(k))))
	}
	co, err := builder.Build()
	if err != nil {
		return err
	}
	return cp.cnt.Update("root", co...)
}

func ChartsFn(cmd *cobra.Command, args []string) error {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	c, err := container.New(t, container.ID("root"))
	if err != nil {
		panic(err)
	}

	cp := &ChartProcessor{
		cnt:    c,
		series: map[string]*ChartSeries{},
	}

	DataProcessorStart(cp)

	ctx, cancel := context.WithCancel(context.Background())
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			cancel()
		}
	}

	const redrawInterval = 250 * time.Millisecond
	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		panic(err)
	}

	return nil
}
