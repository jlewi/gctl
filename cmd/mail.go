package cmd

import (
	"context"
	"fmt"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"os"

	"github.com/jlewi/gctl/gsuite"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewMailCmd adds commands to deal with gmail
func NewMailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "mail",
	}

	cmd.AddCommand(NewMailSearchCmd())
	cmd.AddCommand(NewMailGetCmd())
	return cmd
}

func NewMailSearchCmd() *cobra.Command {
	var maxResults int64
	var pageToken string
	cmd := &cobra.Command{
		Use:  "search <query>",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				app := gsuite.NewApp(os.Stdout)
				if err := app.LoadConfig(nil); err != nil {
					return err
				}

				if err := app.SetupTokenSource(); err != nil {
					return err
				}

				inbox, err := gsuite.NewInbox(*app.Config, app.TS)
				if err != nil {
					return err
				}

				query := args[0]
				results, err := inbox.Search(context.Background(), query, maxResults, pageToken)

				log := zapr.NewLogger(zap.L())
				if _, err := fmt.Fprintf(app.Out, "%s\n", helpers.PrettyString(results)); err != nil {
					log.Error(err, "Failed to write results to output")
				}

				if err != nil {
					return errors.Wrapf(err, "Error searching gmail")
				}

				return nil
			}()

			if err != nil {
				fmt.Printf("Failed to search mail;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().Int64VarP(&maxResults, "max-results", "m", 25, "Maximum number of results to return")
	cmd.Flags().StringVarP(&pageToken, "page-token", "p", "", "The page token to use to fetch the next page of results")
	return cmd
}

func NewMailGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "get <message id>",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				app := gsuite.NewApp(os.Stdout)
				if err := app.LoadConfig(nil); err != nil {
					return err
				}

				if err := app.SetupTokenSource(); err != nil {
					return err
				}

				inbox, err := gsuite.NewInbox(*app.Config, app.TS)
				if err != nil {
					return err
				}

				messageID := args[0]
				results, err := inbox.GetMessage(context.Background(), messageID)

				log := zapr.NewLogger(zap.L())
				if _, err := fmt.Fprintf(app.Out, "%s\n", helpers.PrettyString(results)); err != nil {
					log.Error(err, "Failed to write results to output")
				}

				if err != nil {
					return errors.Wrapf(err, "Error getting message")
				}

				return nil
			}()

			if err != nil {
				fmt.Printf("Failed to get message;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
