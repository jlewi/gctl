package cmd

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/jlewi/gctl/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewConfigCmd adds commands to deal with configuration
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "config",
	}

	cmd.AddCommand(NewGetConfigCmd())
	cmd.AddCommand(NewSetConfigCmd())
	return cmd
}

// NewSetConfigCmd sets a key value pair in the configuration
func NewSetConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "set <name>=<value>",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			err := func() error {
				if err := config.InitViper(cmd); err != nil {
					return err
				}

				pieces := strings.Split(args[0], "=")
				cfgName := pieces[0]

				var fConfig *config.Config
				if len(pieces) < 2 {
					return errors.New("Invalid usage; set expects an argument in the form <NAME>=<VALUE>")
				}
				cfgValue := pieces[1]
				viper.Set(cfgName, cfgValue)
				fConfig = config.GetConfig()

				file := viper.ConfigFileUsed()
				if file == "" {
					file = config.DefaultConfigFile()
				}
				// Persist the configuration
				return fConfig.Write(file)
			}()

			if err != nil {
				fmt.Printf("Failed to set configuration;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

// NewGetConfigCmd  prints out the configuration
func NewGetConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: fmt.Sprintf("Dump %s configuration as YAML", appName),
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				if err := config.InitViper(cmd); err != nil {
					return err
				}
				fConfig := config.GetConfig()

				if err := yaml.NewEncoder(os.Stdout).Encode(fConfig); err != nil {
					return err
				}

				return nil
			}()

			if err != nil {
				fmt.Printf("Failed to get configuration;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
