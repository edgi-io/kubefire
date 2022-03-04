package cluster

import (
	"github.com/edgi-io/kubefire/internal/di"
	"github.com/edgi-io/kubefire/internal/validate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var forceDeleteCluster bool

var deleteCmd = &cobra.Command{
	Use:     "delete [name, ...]",
	Aliases: []string{"rm", "del"},
	Short:   "Deletes clusters",
	Args:    validate.MinimumArgs("name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !forceDeleteCluster {
			return validate.CheckClusterExist(args[0])
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, n := range args {
			if err := di.ClusterManager().Delete(n, forceDeleteCluster); err != nil {
				return errors.WithMessagef(err, "failed to delete cluster (%s)", n)
			}
		}

		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceDeleteCluster, "force", "f", false, "Force to delete")
}
