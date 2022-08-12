package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ssbeatty/jsonschema"
	"os"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "plugin schema",
	Args:  cobra.MinimumNArgs(0),
	Long:  `get plugin schema`,
	Run: func(cmd *cobra.Command, args []string) {
		getPluginSchema(args)
	},
}

func getPluginSchema(args []string) {
	ref := jsonschema.Reflector{DoNotReference: true}

	schema := ref.Reflect(&Params{})

	out, err := json.Marshal(schema)
	if err != nil {
		fmt.Fprintf(os.Stderr, "生成schema出错, err: %v", err)
		os.Exit(-1)
	}

	fmt.Fprintf(os.Stdout, string(out))
}
