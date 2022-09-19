package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var (
	clients string
	params  string
)

type Params struct {
	VNCPassWord string `json:"vnc_pass_word" jsonschema:"required=true" jsonschema_description:"VNC密码"`
	VNCDisplay  int    `json:"vnc_display"  jsonschema:"required=true,default=0" jsonschema_description:"VNC Display Port, 默认: 0"`
	Auth        string `json:"auth" jsonschema:"required=true,default=guess" jsonschema_description:"VNC Auth, Default guess"`
}

type ClientConfig struct {
	Host       string `json:"host"`
	User       string `json:"user"`
	Password   string `json:"password"`
	Passphrase string `json:"passphrase"`
	KeyBytes   []byte `json:"key_bytes"`
	Port       int    `json:"port"`
}

var rootCmd = &cobra.Command{
	Use:   "vnc_install",
	Short: "install x11vnc",
	Long:  `install x11vnc without network!`,
}

func init() {
	rootCmd.AddCommand(nameCmd)
	rootCmd.AddCommand(descCmd)
	rootCmd.AddCommand(schemaCmd)
	rootCmd.AddCommand(execCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
