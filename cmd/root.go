package cmd

import (
	"fmt"
	"os"

	"github.com/jlewi/gctl/config"
	"github.com/spf13/cobra"
)

const (
	appName = "gctl"
)

func NewRootCmd() *cobra.Command {
	var cfgFile string
	var level string
	rootCmd := &cobra.Command{
		Short: appName,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, config.ConfigFlagName, "", fmt.Sprintf("config file (default is $HOME/.%s/config.yaml)", appName))
	rootCmd.PersistentFlags().StringVarP(&level, config.LevelFlagName, "", "info", "The logging level.")

	rootCmd.AddCommand(NewConfigCmd())
	rootCmd.AddCommand(NewMailCmd())
	rootCmd.AddCommand(NewVersionCmd(appName, os.Stdout))
	return rootCmd
}
