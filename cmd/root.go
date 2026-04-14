package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgPath string

var rootCmd = &cobra.Command{
	Use:   "bruwrapper",
	Short: "A CLI wrapper for Bruno API client",
	Long:  "bruwrapper wraps the Bruno CLI (bru) providing better UX for ad-hoc API consumption.",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "Config file path (default ~/.bruwrapper.yaml)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
