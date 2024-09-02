package cmd

import (
	"fmt"
	"github.com/jlewi/gctl/gsuite"
	"github.com/jlewi/monogo/helpers"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"os"
)

func NewDriveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "drive",
	}

	cmd.AddCommand(NewImportCmd())
	cmd.AddCommand(NewSearchCmd())
	return cmd
}

func NewImportCmd() *cobra.Command {
	var path string
	var title string
	var folderID string
	cmd := &cobra.Command{
		Use: "import",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				app := gsuite.NewApp(os.Stdout)
				if err := app.LoadConfig(nil); err != nil {
					return err
				}

				if err := app.SetupTokenSource(); err != nil {
					return err
				}

				d, err := gsuite.NewDrive(*app.Config, app.TS)
				if err != nil {
					return err
				}

				url, err := d.ImportToGoogleDoc(context.Background(), path, title, folderID)

				if err != nil {
					fmt.Fprintf(app.Out, "Error importing the document: %v\n", err)
					return err
				} else {
					fmt.Fprintf(app.Out, "Successfully imported document to Google Doc:\n%s\n", url)
				}

				return nil
			}()

			if err != nil {
				fmt.Printf("Failed to import document;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&title, "title", "t", "", "The title for the document")
	cmd.Flags().StringVarP(&path, "file", "f", "", "The file to import")
	cmd.Flags().StringVarP(&folderID, "folder-id", "p", "", "The id of the folder to import the document to")
	cmd.MarkFlagRequired("file")
	return cmd
}

func NewSearchCmd() *cobra.Command {
	var maxResults int64
	var pageToken string
	var query string
	cmd := &cobra.Command{
		Use: "search",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				app := gsuite.NewApp(os.Stdout)
				if err := app.LoadConfig(nil); err != nil {
					return err
				}

				if err := app.SetupTokenSource(); err != nil {
					return err
				}

				d, err := gsuite.NewDrive(*app.Config, app.TS)
				if err != nil {
					return err
				}

				results, err := d.Search(context.Background(), query, maxResults, pageToken)

				if err != nil {
					fmt.Fprintf(app.Out, "Error searching Google Drive: %v\n", err)
					return err
				}
				fmt.Fprintf(app.Out, helpers.PrettyString(results))

				return nil
			}()

			if err != nil {
				fmt.Printf("Failed to search drive;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().Int64VarP(&maxResults, "max-results", "m", 25, "Maximum number of results to return")
	cmd.Flags().StringVarP(&pageToken, "page-token", "p", "", "The page token to use to fetch the next page of results")
	cmd.Flags().StringVarP(&query, "query", "q", "", "The query to run")
	cmd.MarkFlagRequired("query")
	return cmd
}
