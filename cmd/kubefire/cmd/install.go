package cmd

import (
	"github.com/edgi-io/kubefire/internal/config"
	"github.com/edgi-io/kubefire/pkg/script"
	"github.com/spf13/cobra"
	"os/exec"
)

var forceDownload bool

var InstallCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"in"},
	Short:   "Installs or updates prerequisites",
	PreRun: func(cmd *cobra.Command, args []string) {
		if !forceDownload {
			forceDownload = !config.IsReleasedTagVersion(config.TagVersion)
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// download install script
		scripts := []script.Type{
			script.InstallPrerequisites,
		}

		for _, s := range scripts {
			if err := script.Download(s, config.TagVersion, forceDownload); err != nil {
				return err
			}

			if err := script.Run(s, config.TagVersion, createSetupInstallCommandEnvsFunc()); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	flags := InstallCmd.Flags()
	flags.BoolVarP(&forceDownload, "force", "f", false, "Force to install")
}

func createSetupInstallCommandEnvsFunc() func(cmd *exec.Cmd) error {
	return func(cmd *exec.Cmd) error {
		cmd.Env = append(
			cmd.Env,
			config.ExpectedPrerequisiteVersionsEnvVars()...,
		)

		return nil
	}
}
