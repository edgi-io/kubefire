package main

import (
	"fmt"
	"github.com/edgi-io/kubefire/cmd/kubefire/cmd"
	"github.com/edgi-io/kubefire/cmd/kubefire/cmd/cache"
	"github.com/edgi-io/kubefire/cmd/kubefire/cmd/cluster"
	"github.com/edgi-io/kubefire/cmd/kubefire/cmd/kubeconfig"
	"github.com/edgi-io/kubefire/cmd/kubefire/cmd/node"
	"github.com/edgi-io/kubefire/internal/config"
	"github.com/edgi-io/kubefire/internal/di"
	"github.com/edgi-io/kubefire/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"path"
	"runtime"
)

var rootCmd = &cobra.Command{
	Use:           "kubefire",
	Short:         "KubeFire, creates and manages Kubernetes clusters using FireCracker microVMs",
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		di.DelayInit(false)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&config.LogLevel, "log-level", "l", logrus.InfoLevel.String(), util.FlagsValuesUsage("log level", logrus.AllLevels))
	rootCmd.PersistentFlags().StringVarP(&config.GithubToken, "github-token", "t", "", "GIthub Personal Access Token used to query repo release info")
}

func initConfig() {
	level, _ := logrus.ParseLevel(config.LogLevel)
	logrus.SetLevel(level)

	formatter := &logrus.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			filename := path.Base(frame.File)
			return fmt.Sprintf("%s()", frame.Function), fmt.Sprintf("%12s:%-4d", filename, frame.Line)
		},
	}
	logrus.SetFormatter(formatter)

	if level >= logrus.TraceLevel {
		logrus.SetReportCaller(true)
	}
}

func main() {
	cmds := []*cobra.Command{
		cmd.VersionCmd,
		cmd.InstallCmd,
		cmd.UninstallCmd,
		cmd.InfoCmd,
		cmd.ImageCmd,
		kubeconfig.Cmd,
		cluster.Cmd,
		node.Cmd,
		cache.Cmd,
	}

	for _, c := range cmds {
		rootCmd.AddCommand(c)
	}

	if err := rootCmd.Execute(); err != nil {
		logrus.Tracef("%+v", err)
		logrus.WithError(err).Fatalf("failed to run kubefire")
	}
}
