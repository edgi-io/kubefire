package cluster

import (
	"fmt"
	"github.com/edgi-io/kubefire/internal/di"
	"github.com/edgi-io/kubefire/internal/validate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	showKubeconfigPathOnly bool
)

var envCmd = &cobra.Command{
	Use:   "env [name]",
	Short: "Prints environment values of cluster",
	Args:  validate.OneArg("cluster name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckClusterExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cluster, err := di.ConfigManager().GetCluster(name)
		if err != nil {
			return errors.WithMessagef(err, "failed to get cluster (%s) config", name)
		}

		if showKubeconfigPathOnly {
			fmt.Print(cluster.LocalKubeConfig())
		} else {
			fmt.Printf("KUBECONFIG=%s\n", cluster.LocalKubeConfig())
		}

		return nil
	},
}

func init() {
	envCmd.Flags().BoolVar(&showKubeconfigPathOnly, "path-only", false, "Show kubeconfig path only")
}
