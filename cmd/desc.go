package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var descCmd = &cobra.Command{
	Use:   "desc",
	Short: "plugin desc",
	Args:  cobra.MinimumNArgs(0),
	Long:  `get plugin desc`,
	Run: func(cmd *cobra.Command, args []string) {
		getPluginDesc(args)
	},
}

func getPluginDesc(args []string) {
	_, err := fmt.Fprintf(os.Stdout, "离线在目标机器安装X11VNC")
	if err != nil {
		os.Exit(-1)
	}
}
