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
	"fmt"
	"os"
	"text/template"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/snappi/gosnappi/otg"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	METRIC_PORT    = "port"
	METRIC_FLOW    = "flow"
	COUNTER_FRAMES = "frames"
	COUNTER_BYTES  = "bytes"
	COUNTER_PPS    = "pps"
	COUNTER_TPUT   = "tput"
	COUNTER_KBPS   = "Kbps"
	COUNTER_MBPS   = "Mbps"
	COUNTER_GBPS   = "Gbps"
	COUNTER_TBPS   = "Tbps"
)

var transformMetrics string      // Metrics type to report: "port" for PortMetrics, "flow" for FlowMetrics
var transformCounters string     // Metric counters to transform:  "frames" for frame count,  "bytes" for byte count,  "pps" for frame rate", "tput" for byte rate)
var transformTemplateFile string // Go template file for transform

// transformCmd represents the transform command
var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Transform raw OTG metrics into a format suitable for further processing",
	Long: `
Transform raw OTG metrics into a format suitable for further processing.

If no parameters is provided, transform validates input for a match with
OTG MetricsResponse data structure, and if matched, outputs it as is.

For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		var templatebytes []byte
		var template string
		var err error

		if transformTemplateFile != "" { // Read template from file
			templatebytes, err = os.ReadFile(transformTemplateFile)
			if err != nil {
				log.Fatal(err)
			}
			template = string(templatebytes)
		} else { // Use built-in templates
			switch transformMetrics {
			case METRIC_PORT:
				switch transformCounters {
				case COUNTER_FRAMES:
					template = otgTemplatePortMetricFrames
				case COUNTER_BYTES:
					template = otgTemplatePortMetricBytes
				case COUNTER_PPS:
					template = otgTemplatePortMetricFrameRate
				case COUNTER_TPUT:
					template = otgTemplatePortMetricByteRate
				case COUNTER_KBPS:
					template = otgTemplatePortMetricBitRate
				case COUNTER_MBPS:
					template = otgTemplatePortMetricBitRate
				case COUNTER_GBPS:
					template = otgTemplatePortMetricBitRate
				case COUNTER_TBPS:
					template = otgTemplatePortMetricBitRate
				case "":
					template = otgTemplatePortMetricFrames
				default:
					log.Fatalf("Unsupported metrics counters requested: %s", transformCounters)
				}
			case METRIC_FLOW:
				switch transformCounters {
				case COUNTER_FRAMES:
					template = otgTemplateFlowMetricFrames
				case COUNTER_BYTES:
					template = otgTemplateFlowMetricBytes
				case COUNTER_PPS:
					template = otgTemplateFlowMetricFrameRate
				case COUNTER_TPUT:
					log.Warning("Flow throughput is a calculated metric which uses average frame size over a duration of the flow transmission.")
					template = otgTemplateFlowMetricByteRate
				case "":
					template = otgTemplateFlowMetricFrames
				default:
					log.Fatalf("Unsupported metrics counters requested: %s", transformCounters)
				}
			default:
				template = otgTemplateMetricResponsePassThrough
			}
		}

		transformStdInWithTemplate(template)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		setLogLevel(cmd, logLevel)
		switch transformMetrics {
		case METRIC_PORT:
		case METRIC_FLOW:
		case "": // this would mean --metrics was not defined, will use passthrough mode
		default:
			log.Fatalf("Unsupported metrics type requested: %s", transformMetrics)
		}
		switch transformCounters {
		case COUNTER_FRAMES:
		case COUNTER_BYTES:
		case COUNTER_PPS:
		case COUNTER_TPUT:
		case COUNTER_KBPS:
		case COUNTER_MBPS:
		case COUNTER_GBPS:
		case COUNTER_TBPS:
		case "":
		default:
			log.Fatalf("Unsupported metrics counter requested: %s", transformCounters)
		}
		return nil
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
	transformCmd.Flags().StringVarP(&transformTemplateFile, "file", "f", "", "Go template file for transform")
	transformCmd.Flags().StringVarP(&transformMetrics, "metrics", "m", "", fmt.Sprintf("Metrics type to transform:\n  \"%s\" for PortMetrics\n  \"%s\" for FlowMetrics\n", METRIC_PORT, METRIC_FLOW))
	transformCmd.MarkFlagsMutuallyExclusive("metrics", "file") // either use parameters to control transformation, or provide a template file
	transformCmd.Flags().StringVarP(&transformCounters, "counters", "c", "", fmt.Sprintf("Metric counters to transform:\n  \"%s\" for frame count (default)\n  \"%s\" for byte count\n  \"%s\" for frame rate, in packets per second\n  \"%s\" for throughput, in bytes per second (calculated for FlowMetrics)\n  \"%s | %s | %s\" for throughput, in bits per second (PortMetrics only)", COUNTER_FRAMES, COUNTER_BYTES, COUNTER_PPS, COUNTER_TPUT, COUNTER_KBPS, COUNTER_MBPS, COUNTER_GBPS))
	transformCmd.MarkFlagsMutuallyExclusive("counters", "file") // either use parameters to control transformation, or provide a template file
}

func transformStdInWithTemplate(t string) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		text := scanner.Text()

		mr := gosnappi.NewMetricsResponse()
		err := mr.FromJson(text)
		if err != nil {
			log.Fatal(err)
		}
		transformMetricsResponse(mr, t)
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
			"counterPrintf": func(f string, c int64) string {
				return fmt.Sprintf(f, c)
			},
			"ratePrintf": func(f string, c float32) string {
				return fmt.Sprintf(f, c)
			},
			"rateToBps": func(f string, c float32) float32 {
				return convertByteRateToBps(f, c)
			},
			"byteRateToBits": func(c float32) float32 {
				return convertByteRateToBps(transformCounters, c)
			},
			"frameRateToByteRate": func(frames int64, bytes int64, rate float32) float32 {
				return convertFrameRateToByteRate(frames, bytes, rate)
			},
			"rateMetric": func(d string) string {
				return rateMetricName(transformCounters, d)
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

func convertByteRateToBps(f string, c float32) float32 {
	switch f {
	case "bps":
		c = c * 8
	case "Kbps":
		c = c * 8 / 1000
	case "Mbps":
		c = c * 8 / (1000 * 1000)
	case "Gbps":
		c = c * 8 / (1000 * 1000 * 1000)
	case "Tbps":
		c = c * 8 / (1000 * 1000 * 1000 * 1000)
	default:
		log.Fatalf("Unsupported format string for byteRateToBits: %s", f)
	}
	return c
}

func convertFrameRateToByteRate(frames int64, bytes int64, rate float32) float32 {
	if frames == 0 || bytes == 0 || rate == 0.0 {
		return 0
	}
	frameAvgSize := float32(bytes / frames)
	return float32(frameAvgSize) * rate
}

func rateMetricName(n string, d string) string {
	var suffix string
	switch d {
	case "tx":
		suffix = "_tx"
	case "rx":
		suffix = "_rx"
	default:
		log.Fatalf("Unsupported direction for rateMetricName: %s", d)
	}
	if n == "" {
		return transformCounters + suffix
	}
	return n + suffix
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

// built-in templates
const (
	otgTemplateMetricResponsePassThrough = `{{ otgMetricsResponseToJson . }}
`
	otgTemplatePortMetricFrames = `[{{range $i, $p := .PortMetrics}}{{if $i}},{{end}}{"name": "{{ $p.Name }}", "frames_tx": "{{ $p.FramesTx }}", "frames_rx": "{{ $p.FramesRx }}"}{{end}}]
`
	otgTemplatePortMetricBytes = `[{{range $i, $p := .PortMetrics}}{{if $i}},{{end}}{"name": "{{ $p.Name }}", "bytes_tx": "{{ $p.BytesTx }}", "bytes_rx": "{{ $p.BytesRx }}"}{{end}}]
`
	otgTemplatePortMetricFrameRate = `[{{range $i, $p := .PortMetrics}}{{if $i}},{{end}}{"name": "{{ $p.Name }}", "frames_tx_rate": "{{ $p.FramesTxRate }}", "frames_rx_rate": "{{ $p.FramesRxRate }}"}{{end}}]
`
	otgTemplatePortMetricByteRate = `[{{range $i, $p := .PortMetrics}}{{if $i}},{{end}}{"name": "{{ $p.Name }}", "bytes_tx_rate": "{{ ratePrintf "%.0f" $p.BytesTxRate }}", "bytes_rx_rate": "{{ ratePrintf "%.0f" $p.BytesRxRate }}"}{{end}}]
`
	otgTemplatePortMetricBitRate = `[{{range $i, $p := .PortMetrics}}{{if $i}},{{end}}{"name": "{{ $p.Name }}", "{{ rateMetric "tx" }}": "{{ byteRateToBits $p.BytesTxRate | ratePrintf "%.4f" }}", "{{ rateMetric "rx" }}": "{{ byteRateToBits $p.BytesRxRate | ratePrintf "%.4f" }}"}{{end}}]
`
	otgTemplateFlowMetricFrames = `[{{range $i, $f := .FlowMetrics}}{{if $i}},{{end}}{"name": "{{ $f.Name }}", "frames_tx": "{{ $f.FramesTx }}", "frames_rx": "{{ $f.FramesRx }}"}{{end}}]
`
	otgTemplateFlowMetricBytes = `[{{range $i, $f := .FlowMetrics}}{{if $i}},{{end}}{"name": "{{ $f.Name }}", "bytes_tx": "{{ $f.BytesTx }}", "bytes_rx": "{{ $f.BytesRx }}"}{{end}}]
`
	otgTemplateFlowMetricFrameRate = `[{{range $i, $f := .FlowMetrics}}{{if $i}},{{end}}{"name": "{{ $f.Name }}", "frames_tx_rate": "{{ ratePrintf "%.0f" $f.FramesTxRate }}", "frames_rx_rate": "{{ ratePrintf "%.0f" $f.FramesRxRate }}"}{{end}}]
`
	otgTemplateFlowMetricByteRate = `[{{range $i, $f := .FlowMetrics}}{{if $i}},{{end}}{"name": "{{ $f.Name }}", "bytes_tx_rate": "{{ frameRateToByteRate $f.FramesTx $f.BytesTx $f.FramesTxRate }}", "bytes_rx_rate": "{{ frameRateToByteRate $f.FramesRx $f.BytesRx $f.FramesRxRate }}"}{{end}}]
`
)
