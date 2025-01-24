package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "cowg",
	Short: "Controller Wireguard",
	Long: "CLI utility for Wireguard",
}

func init() {
	rootCmd.AddCommand(sshCommand)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
