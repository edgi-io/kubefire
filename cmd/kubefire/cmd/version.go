package cmd

import (
	"fmt"
	"github.com/edgi-io/kubefire/internal/config"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Shows version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nBuild: %s\n", config.TagVersion, config.BuildVersion)
	},
}
