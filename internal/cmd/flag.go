package cmd

import (
	"github.com/edgi-io/kubefire/internal/config"
	"github.com/edgi-io/kubefire/pkg/output"
	"github.com/edgi-io/kubefire/pkg/util"
	"github.com/spf13/cobra"
)

func AddOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&config.Output, "output", "o", string(output.DEFAULT), util.FlagsValuesUsage("output format", output.BuiltinTypes))
}
