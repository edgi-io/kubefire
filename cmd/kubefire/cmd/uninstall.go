package cmd

import (
	"github.com/edgi-io/kubefire/internal/config"
	"github.com/edgi-io/kubefire/pkg/script"
	"github.com/spf13/cobra"
)

var UninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Aliases: []string{"un"},
	Short:   "Uninstalls prerequisites",
	PreRun: func(cmd *cobra.Command, args []string) {
		if !forceDownload {
			forceDownload = !config.IsReleasedTagVersion(config.TagVersion)
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := script.Download(script.UninstallPrerequisites, config.TagVersion, forceDownload); err != nil {
			return err
		}

		if err := script.Run(script.UninstallPrerequisites, config.TagVersion, createSetupInstallCommandEnvsFunc()); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	flags := UninstallCmd.Flags()
	flags.BoolVarP(&forceDownload, "force", "f", false, "Force to uninstall")
}
