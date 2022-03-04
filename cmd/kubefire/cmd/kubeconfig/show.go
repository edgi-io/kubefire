package kubeconfig

import (
	"fmt"
	"github.com/edgi-io/kubefire/internal/config"
	"github.com/edgi-io/kubefire/internal/di"
	"github.com/edgi-io/kubefire/internal/validate"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var showCmd = &cobra.Command{
	Use:     "show [cluster-name]",
	Aliases: []string{"g"},
	Short:   "Shows the kubeconfig of cluster",
	Args:    validate.OneArg("cluster name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		logrus.SetLevel(logrus.ErrorLevel)

		return validate.CheckClusterExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cluster, err := di.ClusterManager().Get(name)
		if err != nil {
			return errors.WithMessagef(err, "failed to get cluster (%s) info", name)
		}

		destDir := os.TempDir()
		defer func() {
			_ = os.RemoveAll(destDir)
		}()

		config.Bootstrapper = cluster.Spec.Bootstrapper
		di.DelayInit(true)

		wd, _ := os.Getwd()
		kubeconfigPath, err := di.Bootstrapper().DownloadKubeConfig(cluster, wd)
		if err != nil {
			return errors.WithMessagef(err, "failed to download kubeconfig of cluster (%s)", name)
		}

		bytes, err := ioutil.ReadFile(kubeconfigPath)
		if err != nil {
			return err
		}

		fmt.Print(string(bytes))

		return nil
	},
}
