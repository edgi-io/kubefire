package cmd

import (
	"github.com/edgi-io/kubefire/internal/di"
	"github.com/edgi-io/kubefire/pkg/bootstrap"
	"github.com/edgi-io/kubefire/pkg/constants"
	"github.com/sirupsen/logrus"
	"strings"
)

type BootstrapperVersionInfo struct {
	Kubeadm  string
	K3s      string
	RKE      string
	RKE2     string
	RANCHERD string
	K0s      string
}

func BootstrapperVersionInfos() *BootstrapperVersionInfo {
	versionsMap := map[string]string{}

	for _, bootstrapperType := range bootstrap.BuiltinTypes {
		bootstrapperType := bootstrapperType
		versionsMap[bootstrapperType] = getVersions(bootstrapperType)
	}

	return &BootstrapperVersionInfo{
		Kubeadm:  versionsMap[constants.KUBEADM],
		K3s:      versionsMap[constants.K3S],
		RKE:      versionsMap[constants.RKE],
		RKE2:     versionsMap[constants.RKE2],
		RANCHERD: versionsMap[constants.RANCHERD],
		K0s:      versionsMap[constants.K0s],
	}
}

func getVersions(bootstrapperType string) string {
	_, supportedVersions, err := bootstrap.GenerateSaveBootstrapperVersions(bootstrapperType, di.ConfigManager())
	if err != nil {
		logrus.WithError(err).Errorln()
		return ""
	}

	var versions []string
	for _, v := range supportedVersions {
		versions = append(versions, v.Display())
	}

	return strings.Join(versions, ",")
}
