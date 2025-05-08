package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
)

// VersionCmd represents the version command
var VersionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v", "V"},
	Short:   "version/v",
	Long:    `Show current version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s %s (%s %s)\n", "sproxy", AppVersion, AppRevision, AppBuildDate)
	},
}

var AppVersion string
var AppRevision string
var AppBuildDate string

var ThisVersion *semver.Version

// Inject
// should be called by init of main
func Inject(version, rev, date string) {
	AppVersion = version
	AppRevision = rev
	AppBuildDate = date
	v, err := semver.NewVersion(AppVersion)
	if err != nil {
		log.Fatalf("Error parsing version: %v", err)
		os.Exit(-1)
	}
	ThisVersion = v
}
