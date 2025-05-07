package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/qiuzhanghua/sproxy/cmd"
	"github.com/spf13/cobra"
)

func init() {
	cmd.Inject(AppVersion, AppRevision, AppBuildDate)
}

func main() {

	var rootCmd = &cobra.Command{
		Use:   "sproxy",
		Short: "sproxy is a secure reverse proxy server",
		Long:  `sproxy is a secure reverse proxy server`,
	}
	rootCmd.AddCommand(cmd.ServeCmd, cmd.VersionCmd)
	cobra.CheckErr(rootCmd.Execute())
}
