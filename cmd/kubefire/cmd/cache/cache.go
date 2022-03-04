package cache

import (
	"github.com/edgi-io/kubefire/internal/di"
	"github.com/edgi-io/kubefire/internal/validate"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "cache",
	Aliases: []string{"c"},
	Short:   "Manages caches",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		di.DelayInit(false)
		return validate.CheckPrerequisites()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	cmds := []*cobra.Command{
		showCmd,
		deleteCmd,
	}

	for _, c := range cmds {
		Cmd.AddCommand(c)
	}
}
