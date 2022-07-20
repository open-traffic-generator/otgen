/*
Copyright © 2022 Open Traffic Generator

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"os"
	"text/template"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/snappi/gosnappi/otg"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

// built-in templates
const (
	otgMetricResponsePassThrough = `{{ otgMetricsResponseToJson . }}
`
	otgPortMetricResponse = `[{{range $i, $p := .PortMetrics}}{{if $i}},{{end}}{"name": "{{ $p.Name }}", "frames_tx": "{{ $p.FramesTx }}", "frames_rx": "{{ $p.FramesRx }}"}{{end}}]
`
)

var transformMetrics string // Metrics type to report: "port" for PortMetrics, "flow" for FlowMetrics

// transformCmd represents the transform command
var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Transform raw OTG metrics into a format suitable for further processing",
	Long: `Transform raw OTG metrics into a format suitable for further processing.

For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		switch transformMetrics {
		case "port":
		case "flow":
		case "":
		default:
			log.Fatalf("Unsupported metrics type requested: %s", transformMetrics)
		}

		readStdIn()
	},
}

func init() {
	rootCmd.AddCommand(transformCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// transformCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// transformCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	transformCmd.Flags().StringVarP(&transformMetrics, "metrics", "m", "", "Metrics type to transform:\n  \"port\" for PortMetrics,\n  \"flow\" for FlowMetrics\n ")
}

func readStdIn() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		text := scanner.Text()

		mr := gosnappi.NewMetricsResponse()
		err := mr.FromJson(text)
		if err != nil {
			log.Fatal(err)
		}
		switch transformMetrics {
		case "port":
			transformMetricsResponse(mr, otgPortMetricResponse)
		case "flow":
		default:
			transformMetricsResponse(mr, otgMetricResponsePassThrough)
		}
	}
}

func transformMetricsResponse(mr gosnappi.MetricsResponse, tmpl string) {
	t, err := template.New("default").
		Funcs(template.FuncMap{
			"otgMetricsResponseToJson": func(r *otg.MetricsResponse) string {
				j, err := otgMetricsResponseToJson(r)
				if err != nil {
					log.Fatal(err)
				}
				return string(j)
			},
		}).
		Parse(tmpl)

	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(os.Stdout, mr.Msg())
	if err != nil {
		log.Fatal(err)
	}
}

func otgMetricsResponseToJson(r *otg.MetricsResponse) ([]byte, error) {
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "",
	}
	return opts.Marshal(r)
}
