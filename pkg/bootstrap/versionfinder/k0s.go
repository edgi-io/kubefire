package versionfinder

import (
	"github.com/edgi-io/kubefire/internal/config"
	"github.com/edgi-io/kubefire/pkg/constants"
	"github.com/edgi-io/kubefire/pkg/data"
	"github.com/edgi-io/kubefire/pkg/util"
	"github.com/sirupsen/logrus"
)

type K0sVersionFinder struct {
	BaseVersionFinder
	githubInfoer *util.GithubInfoer
	owner        string
	repo         string
}

func NewK0sVersionFinder() *K0sVersionFinder {
	return &K0sVersionFinder{
		BaseVersionFinder: BaseVersionFinder{
			constants.K0s,
		},
		githubInfoer: util.NewGithubInfoer(config.GithubToken),
		owner:        "k0sproject",
		repo:         "k0s",
	}
}

func (k *K0sVersionFinder) GetVersionsAfterVersion(afterVersion data.Version) ([]*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Debugln("getting the released versions info")

	return k.githubInfoer.GetVersionsAfterVersion(afterVersion, k.owner, k.repo, data.SupportedMinorVersionCount)
}

func (k *K0sVersionFinder) GetLatestVersion() (*data.Version, error) {
	logrus.WithField("bootstrapper", k.bootstrapperType).Debugln("getting the latest released version info")

	return k.githubInfoer.GetLatestVersion(k.owner, k.repo)
}
