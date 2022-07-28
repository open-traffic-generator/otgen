package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosuri/uilive"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type TableProcessor struct {
	headers  []string
	data     map[string]map[string]string
	terminal *uilive.Writer
}

func (tp *TableProcessor) Layout(input []DataPoint) error {
	fields := map[string]string{}
	for _, p := range input {
		for k := range p {
			if k != NAME_FIELD {
				fields[k] = k
			}
		}
	}

	tp.headers = []string{NAME_FIELD}
	for k := range fields {
		tp.headers = append(tp.headers, k)
	}

	return nil
}

func (tp *TableProcessor) Process(input []DataPoint) error {
	for _, p := range input {
		name := p[NAME_FIELD].(string)
		if _, ok := tp.data[name]; !ok {
			tp.data[name] = map[string]string{}
			for _, h := range tp.headers {
				tp.data[name][h] = ""
			}
		}

		for _, h := range tp.headers {
			tp.data[name][h] = fmt.Sprintf("%v", p[h])
		}
	}
	fmt.Fprintln(tp.terminal, tp.Format())
	// without this sleep uilive terminal does not get refreshed correctly
	time.Sleep(tp.terminal.RefreshInterval)
	return nil
}

func (tp *TableProcessor) Format() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader(tp.headers)

	for _, v := range tp.data {
		line := make([]string, len(tp.headers))
		for i, h := range tp.headers {
			line[i] = v[h]
		}
		table.Append(line)
	}

	table.Render()
	return tableString.String()
}

func TableFn(cmd *cobra.Command, args []string) error {
	tp := &TableProcessor{
		headers:  []string{},
		data:     map[string]map[string]string{},
		terminal: uilive.New(),
	}
	tp.terminal.RefreshInterval = time.Microsecond * 250
	tp.terminal.Start()

	wg := DataProcessorStart(tp)

	wg.Wait()
	tp.terminal.Stop()

	return nil
}
