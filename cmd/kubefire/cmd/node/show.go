package node

import (
	intcmd "github.com/edgi-io/kubefire/internal/cmd"
	"github.com/edgi-io/kubefire/internal/di"
	"github.com/edgi-io/kubefire/internal/validate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Shows node info",
	Args:  validate.OneArg("node name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckNodeExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		node, err := di.NodeManager().GetNode(name)
		if err != nil {
			return errors.WithMessagef(err, "failed to get node (%s) info", name)
		}

		if err := di.Output().Print(node, nil, ""); err != nil {
			return errors.WithMessagef(err, "failed to print output of node (%s)", name)
		}

		return nil
	},
}

func init() {
	intcmd.AddOutputFlag(showCmd)
}
