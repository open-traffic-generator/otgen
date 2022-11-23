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
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	LOG_DEFAULT_LEVEL = "err"
)

var logLevel string // Logging level: error | info | debug

// Create a new instance of the logger
var log = logrus.New()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "otgen",
	Short: "Open Traffic Generator CLI Tool",
	Long: `
Open Traffic Generator CLI Tool.

For more information, go to https://github.com/open-traffic-generator/otgen
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.otgen.yaml)")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log", "", LOG_DEFAULT_LEVEL, "Logging level: err | warn | info | debug")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}

func setLogLevel(cmd *cobra.Command, logLevel string) {

	switch logLevel {
	case "err":
		log.SetLevel(logrus.ErrorLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	default:
		log.Fatalf("Unsupported log level: %s", logLevel)
	}
	log.Debugf("%s: log level set to %s", cmdFullName(cmd), logLevel)
}

func cmdFullName(cmd *cobra.Command) string {
	var name string
	if cmd == nil {
		name = ""
	} else if cmd.Parent() != nil {
		name = cmdFullName(cmd.Parent()) + " " + cmd.Name()
	} else {
		name = cmd.Name()
	}
	return name
}
