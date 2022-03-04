package versionfinder

import (
	"fmt"
	"github.com/edgi-io/kubefire/internal/config"
	"github.com/edgi-io/kubefire/pkg/constants"
	"github.com/edgi-io/kubefire/pkg/data"
	"github.com/edgi-io/kubefire/pkg/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

const KubeStableVersionUrl = "https://storage.googleapis.com/kubernetes-release/release/stable.txt"

type KubeadmVersionFinder struct {
	BaseVersionFinder

	githubInfoer *util.GithubInfoer
	owner        string
	repo         string
}

func NewKubeadmVersionFinder() *KubeadmVersionFinder {
	return &KubeadmVersionFinder{
		BaseVersionFinder: BaseVersionFinder{
			constants.KUBEADM,
		},
		githubInfoer: util.NewGithubInfoer(config.GithubToken),
		owner:        "kubernetes",
		repo:         "kubernetes",
	}
}

func (k *KubeadmVersionFinder) GetVersionsAfterVersion(afterVersion data.Version) ([]*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Debugf("getting the released versions info less than/equal to %s", afterVersion.String())

	return k.githubInfoer.GetVersionsAfterVersion(afterVersion, k.owner, k.repo, data.SupportedMinorVersionCount)
}

func (k *KubeadmVersionFinder) GetLatestVersion() (*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Debugln("getting the latest released version info")

	body, _, err := util.HttpGet(KubeStableVersionUrl)
	if err != nil {
		return nil, err
	}

	if ok, _ := regexp.MatchString(`v\d+\.\d+\.\d+`, body); !ok {
		return nil, errors.New(fmt.Sprintf("invalid semantic version: %s", body))
	}

	return data.ParseVersion(strings.TrimSuffix(body, "\n")), nil
}

func (k *KubeadmVersionFinder) GetCritoolVersionsAfterVersion(afterVersion data.Version) ([]*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Infof("getting the cri-tools release versions info less than/equal to %s", afterVersion.String())

	return k.githubInfoer.GetVersionsAfterVersion(afterVersion, "kubernetes-sigs", "cri-tools", data.SupportedMinorVersionCount)
}

func (k *KubeadmVersionFinder) GetKubeReleaseToolLatestVersion(version data.Version) (*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Infoln("getting the latest kube release tool release version info")

	return k.githubInfoer.GetLatestVersion("kubernetes", "release")
}
