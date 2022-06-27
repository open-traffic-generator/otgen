/*
Copyright © 2022 Open Traffic Generator

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions://

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
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

// URL of OTG server API endpoint
var otgURL string
var otgYaml string
var otgJson string

// Create a new instance of the logger
var log = logrus.New()

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Request API endpoint to run OTG model",
	Long: `Request API endpoint to run OTG model.

For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		var otgFile string
		// these are mutually exclusive
		if otgYaml != "" {
			otgFile = otgYaml
		}
		if otgJson != "" {
			otgFile = otgJson
			log.Fatal("JSON import is not implemented yet")
		}
		if otgFile == "" {
			log.Fatal("Stdin for OTG input is not implemented yet")
		}
		api, config := initOTG(otgFile)
		flowMetrics := runTraffic(api, config)
		checkPacketLoss(flowMetrics)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	runCmd.Flags().StringVarP(&otgURL, "api", "", "", "URL of OTG API endpoint, for example https://otg-api-endpoint")
	runCmd.MarkFlagRequired("api")
	runCmd.Flags().StringVarP(&otgYaml, "yaml", "y", "", "OTG mode file, in YAML format. Mutually exclusive with --json. If neither is provided, will use stdin")
	runCmd.Flags().StringVarP(&otgJson, "json", "j", "", "OTG mode file, in JSON format. Mutually exclusive with --yaml. If neither is provided, will use stdin")
	runCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	// Log to file
	file, err := os.OpenFile("otgen.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
}

func initOTG(otgfile string) (gosnappi.GosnappiApi, gosnappi.Config) {
	// Read OTG config
	otgbytes, err := ioutil.ReadFile(otgfile)
	if err != nil {
		log.Fatal(err)
	}

	otg := string(otgbytes)

	// Create a new API handle to make API calls against a traffic generator
	api := gosnappi.NewApi()

	// Set the transport protocol to either HTTP or GRPC
	api.NewHttpTransport().SetLocation(otgURL).SetVerify(false)

	// Create a new traffic configuration that will be set on traffic generator
	config := api.NewConfig()
	config.FromYaml(otg)

	return api, config
}

func runTraffic(api gosnappi.GosnappiApi, config gosnappi.Config) gosnappi.MetricsResponse {
	// push traffic configuration to otgHost
	fmt.Printf("Applying OTG config...")
	res, err := api.SetConfig(config)
	checkResponse(res, err)
	fmt.Printf("ready.\n")

	// start transmitting configured flows
	fmt.Printf("Starting traffic...")
	ts := api.NewTransmitState().SetState(gosnappi.TransmitStateState.START)
	res, err = api.SetTransmitState(ts)
	checkResponse(res, err)
	fmt.Printf("started...")

	trafficETA := calculateTrafficETA(config)
	fmt.Printf("ETA is: %s\n", trafficETA)

	// initialize flow metrics
	mr := api.NewMetricsRequest()
	mr.Flow()
	flowMetrics, err := api.GetMetrics(mr)
	if err != nil {
		log.Fatal(err)
	}

	// launch a routine to periodically pull flow metrics
	var flowMetricsMutex sync.Mutex
	keepPulling := true
	go func() {
		for keepPulling {
			flowMetricsMutex.Lock()
			flowMetrics, err = api.GetMetrics(mr)
			flowMetricsMutex.Unlock()
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// use a waitGroup to track progress of each individual flow
	var wg sync.WaitGroup
	// progress bar indicator
	p := mpb.New(mpb.WithWaitGroup(&wg))
	wg.Add(len(config.Flows().Items()))

	// wait for traffic to stop on each flow or run beyond ETA
	start := time.Now()
	for _, f := range config.Flows().Items() {
		// decorate progress bars
		barname := fmt.Sprintf("%s:", f.Name())
		bar := p.AddBar(int64(f.Duration().FixedPackets().Packets()),
			mpb.PrependDecorators(
				// simple name decorator
				decor.Name(barname, decor.WCSyncSpace),
				decor.Percentage(decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				decor.Counters(0, "packets Rx: %d / %d", decor.WCSyncSpaceR),
			),
		)

		go func(f gosnappi.Flow) {
			defer flowMetricsMutex.Unlock()
			defer wg.Done()
			for {
				flowMetricsMutex.Lock()
				for _, fm := range flowMetrics.FlowMetrics().Items() {
					if fm.Name() == f.Name() {
						bar.IncrInt64(fm.FramesRx() - bar.Current())
						checkResponse(fm, err)
						if fm.Transmit() == gosnappi.FlowMetricTransmit.STOPPED {
							bar.Abort(false)
							return
						}
						if trafficETA*2 < time.Since(start) {
							bar.Abort(false)
							log.Printf("Traffic %s has been running twice longer than ETA, forcing to stop", fm.Name())
							return
						}
					}
				}
				flowMetricsMutex.Unlock()
				time.Sleep(500 * time.Millisecond)
			}
		}(f)
	}
	p.Wait()
	keepPulling = false // stop metrics pulling routine

	// stop transmitting traffic
	fmt.Printf("Stopping traffic...")
	ts = api.NewTransmitState().SetState(gosnappi.TransmitStateState.STOP)
	res, err = api.SetTransmitState(ts)
	checkResponse(res, err)
	fmt.Printf("stopped.\n")

	return flowMetrics
}

func calculateTrafficETA(config gosnappi.Config) time.Duration {
	// Initialize packet counts and rates per flow if they were provided as parameters. Calculate ETA
	flowETA := time.Duration(0)
	trafficETA := time.Duration(0)
	for _, f := range config.Flows().Items() {
		pktCountFlow := f.Duration().FixedPackets().Packets()
		ratePPSFlow := f.Rate().Pps()
		// Calculate ETA it will take to transmit the flow
		if ratePPSFlow > 0 {
			flowETA = time.Duration(float64(int64(pktCountFlow)/ratePPSFlow)) * time.Second
		}
		if flowETA > trafficETA {
			trafficETA = flowETA // The longest flow to finish
		}
	}
	return trafficETA
}

// print otg api response content
func checkResponse(res interface{}, err error) {
	if err != nil {
		log.Fatal(err)
	}
	switch v := res.(type) {
	case gosnappi.MetricsResponse:
		for _, fm := range v.FlowMetrics().Items() {
			log.Printf("Traffic stats for %s:\n%s\n", fm.Name(), fm)
		}
	case gosnappi.FlowMetric:
		log.Printf("Traffic stats for %s:\n%s\n", v.Name(), v)
	case gosnappi.ResponseWarning:
		for _, w := range v.Warnings() {
			log.Printf("WARNING:", w)
		}
	default:
		log.Fatal("Unknown response type:", v)
	}
}

// check for packet loss
func checkPacketLoss(flowMetrics gosnappi.MetricsResponse) {
	for _, fm := range flowMetrics.FlowMetrics().Items() {
		if fm.FramesTx() > 0 {
			loss := float32(fm.FramesTx()-fm.FramesRx()) / float32(fm.FramesTx())
			positiveTest := true // default assumption is that we are testing for a flow to succeed
			if positiveTest {    // we expect the flow to succeed
				if loss > 0.01 {
					log.Fatalf("Packet loss was detected for %s! Measured %f per cent", fm.Name(), loss*100)
				}
			} else { // we expect the flow to fail
				if loss < 0.01 {
					log.Fatalf("Packet loss was expected for %s! Measured %f per cent", fm.Name(), loss*100)
				}
			}
		}
	}
}
