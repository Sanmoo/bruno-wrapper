package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/sanmoo/bruwrapper/internal/app"
	"github.com/sanmoo/bruwrapper/internal/core"
	"github.com/spf13/cobra"
)

var (
	runCollection string
	runRequest    string
	runEnv        string
	runVars       []string
	runRaw        bool
	runVerbose    bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a Bruno request",
	Long:  "Run a request from a Bruno collection. Opens interactive selection if -c and -r are not provided.",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, catalog, runner, presenter, selector, err := wireUp()
		if err != nil {
			return err
		}
		runApp := app.NewRunApp(catalog, runner, presenter, selector)

		var vars []core.Variable
		for _, v := range runVars {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid variable format %q, expected key=value", v)
			}
			vars = append(vars, core.Variable{Key: parts[0], Value: parts[1]})
		}

		return runApp.Run(context.Background(), app.RunParams{
			CollectionName: runCollection,
			RequestName:    runRequest,
			Env:            runEnv,
			Variables:      vars,
			Raw:            runRaw,
			Verbose:        runVerbose,
		})
	},
}

func init() {
	runCmd.Flags().StringVarP(&runCollection, "collection", "c", "", "Collection name")
	runCmd.Flags().StringVarP(&runRequest, "request", "r", "", "Request name")
	runCmd.Flags().StringVarP(&runEnv, "env", "e", "", "Environment name")
	runCmd.Flags().StringArrayVarP(&runVars, "var", "v", nil, "Variable override (key=value, repeatable)")
	runCmd.Flags().BoolVar(&runRaw, "raw", false, "Output raw response without pretty-print")
	runCmd.Flags().BoolVar(&runVerbose, "verbose", false, "Show request and response headers")
	rootCmd.AddCommand(runCmd)
}
