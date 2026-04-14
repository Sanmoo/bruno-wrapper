package cmd

import (
	"github.com/sanmoo/bruwrapper/internal/app"
	"github.com/spf13/cobra"
)

var listCollection string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List collections or requests",
	Long:  "List available collections, or list requests in a specific collection.",
	RunE: func(cmd *cobra.Command, args []string) error {
		catalog, presenter, err := wireCatalogAndPresenter()
		if err != nil {
			return err
		}
		listApp := app.NewListApp(catalog, presenter)

		if listCollection == "" {
			return listApp.ListCollections()
		}
		return listApp.ListRequests(listCollection)
	},
}

func init() {
	listCmd.Flags().StringVarP(&listCollection, "collection", "c", "", "Collection name to list requests from")
	rootCmd.AddCommand(listCmd)
}
