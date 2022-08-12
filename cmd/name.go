package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var nameCmd = &cobra.Command{
	Use:   "name",
	Short: "plugin name",
	Args:  cobra.MinimumNArgs(0),
	Long:  `get plugin name`,
	Run: func(cmd *cobra.Command, args []string) {
		getPluginName(args)
	},
}

func getPluginName(args []string) {
	_, err := fmt.Fprintf(os.Stdout, "vnc_install")
	if err != nil {
		os.Exit(-1)
	}
}
