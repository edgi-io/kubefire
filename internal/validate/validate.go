package validate

import (
	"fmt"
	intcmd "github.com/edgi-io/kubefire/internal/cmd"
	"github.com/edgi-io/kubefire/internal/di"
	interr "github.com/edgi-io/kubefire/internal/error"
	"github.com/edgi-io/kubefire/pkg/bootstrap"
	"github.com/edgi-io/kubefire/pkg/constants"
	"github.com/edgi-io/kubefire/pkg/data"
	"github.com/pkg/errors"
	"runtime"
)

func CheckPrerequisites() error {
	if intcmd.CurrentPrerequisitesInfos().Matched() {
		return nil
	}

	return errors.WithMessage(interr.IncorrectRequiredPrerequisitesError, "check your installed prerequisites by `kubefire info`, then install/update via 'kubefire install'")
}

func CheckClusterExist(name string) error {
	_, err := di.ConfigManager().GetCluster(name)
	if err != nil {
		return errors.WithMessage(interr.ClusterNotFoundError, Field("cluster", name))
	}

	return nil
}

func CheckNodeExist(name string) error {
	if _, err := di.NodeManager().GetNode(name); err != nil {
		return errors.WithMessage(interr.NodeNotFoundError, Field("node", name))
	}

	return nil
}

func CheckClusterVersion(version string) error {
	if version == "" {
		return nil
	}

	if data.ParseVersion(version) == nil {
		return errors.WithMessage(interr.ClusterVersionInvalidError, Field("version", version))
	}

	return nil
}

func CheckBootstrapperType(bootstrapper string) error {
	if !bootstrap.IsValid(bootstrapper) {
		return errors.WithMessage(interr.BootstrapperNotFoundError, Field("bootstrapper", bootstrapper))
	}

	if runtime.GOARCH == "arm64" {
		if bootstrapper == constants.KUBEADM || bootstrapper == constants.RKE {
			return errors.WithMessage(interr.BootstrapperNotSupportError, Field("bootstrapper", bootstrapper))
		}
	}

	return nil
}

func Field(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}
