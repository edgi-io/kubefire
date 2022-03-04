package node

import (
	"github.com/edgi-io/kubefire/internal/di"
	"github.com/edgi-io/kubefire/internal/validate"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Starts node",
	Args:  validate.OneArg("node name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckNodeExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return startNode(args[0])
	},
}

func startNode(name string) error {
	node, _ := di.NodeManager().GetNode(name)

	if node.Status.Running {
		logrus.WithField("node", node.Name).Infoln("node is already running")
		return nil
	}

	if err := di.NodeManager().StartNode(name); err != nil {
		return errors.WithMessagef(err, "failed to start node (%s)", name)
	}

	return nil
}
