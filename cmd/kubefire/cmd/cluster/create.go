package cluster

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/edgi-io/kubefire/internal/config"
	"github.com/edgi-io/kubefire/internal/di"
	"github.com/edgi-io/kubefire/internal/validate"
	"github.com/edgi-io/kubefire/pkg/bootstrap"
	pkgconfig "github.com/edgi-io/kubefire/pkg/config"
	"github.com/edgi-io/kubefire/pkg/data"
	"github.com/edgi-io/kubefire/pkg/util"
	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

var (
	cluster      = pkgconfig.NewDefaultCluster()
	noStart      bool
	noCache      bool
	extraOptions string
	configFile   string
)

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Creates cluster",
	Args:  validate.OneArg("cluster name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if configFile != "" {
			bytes, err := ioutil.ReadFile(configFile)
			if err != nil {
				return errors.WithMessage(err, fmt.Sprintf("failed to get the cluster config file %s", configFile))
			}

			if err := yaml.Unmarshal(bytes, cluster); err != nil {
				return errors.WithMessage(err, fmt.Sprintf("failed to parse the cluster config file %s", configFile))
			}
		}

		if err := validate.CheckBootstrapperType(cluster.Bootstrapper); err != nil {
			return err
		}

		reinitDI := config.Bootstrapper != cluster.Bootstrapper
		config.Bootstrapper = cluster.Bootstrapper
		di.DelayInit(reinitDI)

		if err := validate.CheckClusterVersion(cluster.Version); err != nil {
			return err
		}

		if noCache {
			_ = di.ConfigManager().DeleteBootstrapperVersions(pkgconfig.NewBootstrapperVersion(cluster.Bootstrapper, ""))
		}
		if _, _, err := bootstrap.GenerateSaveBootstrapperVersions(config.Bootstrapper, di.ConfigManager()); err != nil {
			return err
		}

		version, err := correctClusterVersion(cluster.Version)
		if err != nil {
			return err
		}
		cluster.Version = version

		cluster.UpdateExtraOptions(extraOptions)

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster.Name = args[0]

		if forceDeleteCluster {
			_ = di.ClusterManager().Delete(cluster.Name, true)
		}

		if err := di.ClusterManager().Init(cluster); err != nil {
			return errors.WithMessagef(err, "failed to init cluster (%s)", cluster.Name)
		}

		if err := di.ClusterManager().Create(cluster.Name, !noStart); err != nil {
			return errors.WithMessagef(err, "failed to create cluster (%s)", cluster.Name)
		}

		if !noStart {
			if err := deployCluster(cluster.Name); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	flags := createCmd.Flags()

	flags.StringVarP(&cluster.Bootstrapper, "bootstrapper", "b", cluster.Bootstrapper, util.FlagsValuesUsage("Bootstrapper type", bootstrap.BuiltinTypes))
	flags.StringVarP(&cluster.Pubkey, "pubkey", "k", "", "Public key")
	flags.StringVarP(&cluster.Version, "version", "v", "", "Version of Kubernetes supported by bootstrapper (ex: v1.18, v1.18.8, empty)")
	flags.StringVarP(&cluster.Image, "image", "i", cluster.Image, "Rootfs container image")
	flags.StringVar(&cluster.KernelImage, "kernel-image", cluster.KernelImage, "Kernel container image")
	flags.StringVar(&cluster.KernelArgs, "kernel-args", cluster.KernelArgs, "Kernel arguments")
	flags.StringVarP(&extraOptions, "extra-options", "o", "", "Extra options (ex: key=value,...) for bootstrapper")

	flags.IntVar(&cluster.Master.Count, "master-count", cluster.Master.Count, "Count of master node")
	flags.IntVar(&cluster.Master.Cpus, "master-cpu", cluster.Master.Cpus, "CPUs of master node")
	flags.StringVar(&cluster.Master.Memory, "master-memory", cluster.Master.Memory, "Memory of master node")
	flags.StringVar(&cluster.Master.DiskSize, "master-size", cluster.Master.DiskSize, "Disk size of master node")

	flags.IntVar(&cluster.Worker.Count, "worker-count", cluster.Worker.Count, "Count of worker node")
	flags.IntVar(&cluster.Worker.Cpus, "worker-cpu", cluster.Worker.Cpus, "CPUs of worker node")
	flags.StringVar(&cluster.Worker.Memory, "worker-memory", cluster.Worker.Memory, "Memory of worker node")
	flags.StringVar(&cluster.Worker.DiskSize, "worker-size", cluster.Worker.DiskSize, "Disk size of worker node")
	flags.StringVarP(&configFile, "config", "c", "", "Cluster configuration file (ex: use 'config-template' command to generate the default cluster config)")

	flags.BoolVarP(&forceDeleteCluster, "force", "f", false, "Force to recreate if the cluster exists")
	flags.BoolVar(&noCache, "no-cache", false, "Forget caches")
	flags.BoolVar(&noStart, "no-start", false, "Don't start nodes")
}

func deployCluster(name string) error {
	cluster, err := di.ClusterManager().Get(name)
	if err != nil {
		return errors.WithMessagef(err, "failed to get cluster (%s) before bootstrapping", name)
	}

	err = di.Bootstrapper().Deploy(
		cluster,
		func() error {
			return di.Bootstrapper().Prepare(cluster, forceDeleteCluster)
		},
	)
	if err != nil {
		return errors.WithMessagef(err, "failed to deploy cluster (%s)", cluster.Name)
	}

	cluster.Spec.Deployed = true
	if err := di.ConfigManager().SaveCluster(&cluster.Spec); err != nil {
		return errors.WithMessagef(err, "failed to mark the cluster (%s) as deployed", cluster.Name)
	}

	_ = retry.Do(func() error {
		if _, err := di.Bootstrapper().DownloadKubeConfig(cluster, ""); err != nil {
			return errors.WithMessagef(err, "failed to download the kubeconfig of cluster (%s)", cluster.Name)
		}

		return nil
	},
		retry.Delay(10*time.Second),
	)

	return nil
}

func correctClusterVersion(version string) (string, error) {
	latestVersion, err := di.VersionFinder().GetLatestVersion()
	if err != nil {
		return "", err
	}

	if version == "" {
		return latestVersion.String(), nil
	}

	bootstrapperVersion := pkgconfig.NewBootstrapperVersion(di.Bootstrapper().Type(), latestVersion.String())
	versions, err := di.ConfigManager().GetBootstrapperVersions(bootstrapperVersion)
	if err != nil {
		return "", err
	}

	majorMinorVersionRegex := regexp.MustCompile(`^v\d+\.\d+$`)
	for _, v := range versions {
		if version == v.Version() {
			return v.Version(), nil
		}

		if strings.HasPrefix(v.Version(), version+".") {
			return v.Version(), nil
		}

		switch {
		case version == v.Version():
			return v.Version(), nil

		case majorMinorVersionRegex.MatchString(version) && strings.HasPrefix(v.Version(), version+"."):
			return v.Version(), nil
		}
	}

	if data.ParseVersion(version) != nil {
		return version, nil
	}

	return "", errors.New("version not found")
}
