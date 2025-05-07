package main

import (
	"github.com/qiuzhanghua/sproxy/cmd"
	"github.com/spf13/cobra"

	_ "github.com/joho/godotenv/autoload"
)

func main() {

	var rootCmd = &cobra.Command{
		Use:   "sproxy",
		Short: "sproxy is a secure reverse proxy server",
		Long:  `sproxy is a secure reverse proxy server`,
	}
	rootCmd.AddCommand(cmd.ServeCmd)
	cobra.CheckErr(rootCmd.Execute())
}
