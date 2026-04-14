package cmd

import (
	"github.com/sanmoo/bruwrapper/internal/app"
	"github.com/spf13/cobra"
)

var (
	showCollection string
	showRequest    string
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show request details before executing",
	Long:  "Show the method, URL, headers, and body of a request without executing it.",
	RunE: func(cmd *cobra.Command, args []string) error {
		catalog, presenter, err := wireCatalogAndPresenter()
		if err != nil {
			return err
		}
		showApp := app.NewShowApp(catalog, presenter)
		return showApp.ShowRequestDetails(showCollection, showRequest)
	},
}

func init() {
	showCmd.Flags().StringVarP(&showCollection, "collection", "c", "", "Collection name (required)")
	showCmd.Flags().StringVarP(&showRequest, "request", "r", "", "Request name (required)")
	showCmd.MarkFlagRequired("collection")
	showCmd.MarkFlagRequired("request")
	rootCmd.AddCommand(showCmd)
}
