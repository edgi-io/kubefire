package cluster

import (
	pkgconfig "github.com/edgi-io/kubefire/pkg/config"
	"github.com/edgi-io/kubefire/pkg/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
)

var configTemplateCmd = &cobra.Command{
	Use:   "config-template [name]",
	Short: "Generates template cluster configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputer := output.NewOutput(output.YAML, os.Stdout)

		if err := outputer.Print(pkgconfig.NewDefaultCluster(), nil, ""); err != nil {
			return errors.New("failed to print output template cluster config")
		}

		return nil
	},
}
