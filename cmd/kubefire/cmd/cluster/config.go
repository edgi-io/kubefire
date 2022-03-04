package cluster

import (
	intcmd "github.com/edgi-io/kubefire/internal/cmd"
	"github.com/edgi-io/kubefire/internal/di"
	"github.com/edgi-io/kubefire/internal/validate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [name]",
	Short: "Shows cluster configuration",
	Args:  validate.OneArg("cluster name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckClusterExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cluster, err := di.ConfigManager().GetCluster(name)
		if err != nil {
			return err
		}

		if err := di.Output().Print(cluster, nil, ""); err != nil {
			return errors.WithMessagef(err, "failed to print output of cluster (%s)", name)
		}

		return nil
	},
}

func init() {
	intcmd.AddOutputFlag(configCmd)
}
