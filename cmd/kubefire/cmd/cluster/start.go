package cluster

import (
	"github.com/edgi-io/kubefire/internal/di"
	"github.com/edgi-io/kubefire/internal/validate"
	"github.com/edgi-io/kubefire/pkg/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Starts cluster",
	Args:  validate.OneArg("cluster name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckClusterExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster, err := startCluster(args[0])
		if err != nil {
			return err
		}

		if cluster.Deployed {
			return nil
		}

		if err := deployCluster(cluster.Name); err != nil {
			return err
		}

		return nil
	},
}

func startCluster(name string) (*config.Cluster, error) {
	if err := di.NodeManager().StartNodes(name); err != nil {
		err := errors.WithMessagef(err, "failed to start all nodes cluster (%s)", name)

		if !forceDeleteCluster {
			return nil, err
		}

		logrus.WithError(err).WithField("node", name).Println()
	}

	cluster, err := di.ConfigManager().GetCluster(name)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}
