package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lrills/helm-unittest/internal/printer"
	"github.com/lrills/helm-unittest/pkg/unittest"
	"github.com/lrills/helm-unittest/pkg/unittest/formatter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// testOptions stres options setup by user in command line
type testOptions struct {
	config         string
	useHelmV3      bool
	colored        bool
	updateSnapshot bool
	withSubChart   bool
	testFiles      []string
	outputFile     string
	outputType     string
}

var testConfig = testOptions{}

var cmd = &cobra.Command{
	Use:   "unittest [flags] CHART [...]",
	Short: "unittest for helm charts",
	Long: `Running chart unittest written in YAML.

This renders your charts locally (without tiller) and
validates the rendered output with the tests defined in
test suite files. Simplest test suite file looks like
below:

---
# CHART_PATH/tests/deployment_test.yaml
suite: test my deployment
templates:
  - deployment.yaml
tests:
  - it: should be a Deployment
    asserts:
      - isKind:
          of: Deployment
---

Put the test files in "tests" directory under your chart
with suffix "_test.yaml", and run:

$ helm unittest my-chart

Or specify the suite files glob path pattern:

$ helm unittest -f 'my-tests/*.yaml' my-chart

Check https://github.com/lrills/helm-unittest for more
details about how to write tests.
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, chartPaths []string) {
		var colored *bool
		if cmd.PersistentFlags().Changed("color") {
			colored = &testConfig.colored
		}

		formatter := formatter.NewFormatter(testConfig.outputFile, testConfig.outputType)
		printer := printer.NewPrinter(os.Stdout, colored)
		runner := unittest.TestRunner{
			Printer:        printer,
			Formatter:      formatter,
			UpdateSnapshot: testConfig.updateSnapshot,
			WithSubChart:   testConfig.withSubChart,
			TestFiles:      testConfig.testFiles,
			OutputFile:     testConfig.outputFile,
		}
		var passed bool

		if !testConfig.useHelmV3 {
			passed = runner.RunV2(chartPaths)
		} else {
			passed = runner.RunV3(chartPaths)
		}

		if !passed {
			os.Exit(1)
		}
	},
}

// main to execute execute unittest command
func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
  cobra.OnInitialize(initConfig)


	cmd.PersistentFlags().StringVar(
		&testConfig.config, "config", "",
		"config ish mea",
	)

	cmd.PersistentFlags().BoolVar(
		&testConfig.colored, "color", false,
		"enforce printing colored output even stdout is not a tty. Set to false to disable color",
	)

	defaultFilePattern := filepath.Join("tests", "*_test.yaml")
	cmd.PersistentFlags().StringArrayVarP(
		&testConfig.testFiles, "file", "f", []string{defaultFilePattern},
		"glob paths of test files location, default to "+defaultFilePattern,
	)

	cmd.PersistentFlags().BoolVarP(
		&testConfig.updateSnapshot, "update-snapshot", "u", false,
		"update the snapshot cached if needed, make sure you review the change before update",
	)

	cmd.PersistentFlags().BoolVarP(
		&testConfig.withSubChart, "with-subchart", "s", true,
		"include tests of the subcharts within `charts` folder",
	)

	cmd.PersistentFlags().StringVarP(
		&testConfig.outputFile, "output-file", "o", "",
		"output-file the file where testresults are written in JUnit format, defaults no output is written to file",
	)

	cmd.PersistentFlags().StringVarP(
		&testConfig.outputType, "output-type", "t", "XUnit",
		"output-type the file-format where testresults are written in, accepted types are (JUnit, NUnit, XUnit)",
	)

	cmd.PersistentFlags().BoolVarP(
		&testConfig.useHelmV3, "helm3", "3", false,
		"parse helm charts as helm3 charts.",
	)

  viper.BindPFlag("helm3", cmd.PersistentFlags().Lookup("helm3"))

	fmt.Println(testConfig)

}



func initConfig() {
  if testConfig.config != "" {
		viper.SetConfigFile(testConfig.config)
		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		} else {
			fmt.Println(err)
		}
	}
}
