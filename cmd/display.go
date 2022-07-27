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
	"fmt"

	"github.com/spf13/cobra"
)

const (
	TYPE_TABLE  = "table"
	TYPE_CHARTS = "charts"
)

var displayType string // type of display to show

// transformCmd represents the transform command
var displayCmd = &cobra.Command{
	Use:   "display [--type charts|table]",
	Short: "Display running test metrics as charts or table",
	Long: `Display running test metrics as charts or table.

For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		switch displayType {
		case TYPE_CHARTS:
			err := chartsFn(cmd, args)
			if err != nil {
				log.Fatal(err)
			}
		case TYPE_TABLE:
			err := tableFn(cmd, args)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if displayType != TYPE_TABLE && displayType != TYPE_CHARTS {
			return fmt.Errorf("unsupported display type: %s", displayType)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(displayCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// transformCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// transformCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	displayCmd.Flags().StringVarP(&displayType, "type", "t", "charts", "charts or table")
}
