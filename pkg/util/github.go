package util

import (
	"context"
	"github.com/edgi-io/kubefire/pkg/data"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"sort"
	"strconv"
)

type GithubInfoer struct {
	client *github.Client
}

func NewGithubInfoer(token string) *GithubInfoer {
	var client *github.Client

	if len(token) != 0 {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	return &GithubInfoer{client: client}
}

func (g *GithubInfoer) GetVersionsAfterVersion(afterVersion data.Version, repoOwner string, repo string, minorVersionCount int) ([]*data.Version, error) {
	var versions []*data.Version

	opt := github.ListOptions{}

done:
	for {
		releases, resp, err := g.client.Repositories.ListReleases(context.Background(), repoOwner, repo, &opt)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		sort.Slice(releases, func(i, j int) bool {
			v1 := data.ParseVersion(releases[i].GetTagName())
			v2 := data.ParseVersion(releases[j].GetTagName())

			if v1 == nil {
				return false
			}

			if v2 == nil {
				return true
			}

			return v1.Compare(v2) >= 0
		})

		for _, release := range releases {
			releaseVersion := data.ParseVersion(release.GetTagName())
			if releaseVersion == nil {
				continue
			}

			if len(releaseVersion.PRERELEASE) > 0 {
				logrus.Debugf("ignored to getting pre-release version: %s", releaseVersion)
				continue
			}

			if releaseVersion.MajorString() != afterVersion.MajorString() {
				afterVersion.Major = data.SubVersionType(strconv.Itoa(afterVersion.Major.ToInt() - 1))
				afterVersion.Minor = data.SubVersionType(strconv.Itoa(100))

				continue
			}

			if releaseVersion.MajorMinorString() != afterVersion.MajorMinorString() {
				if releaseVersion.Minor.ToInt() > afterVersion.Minor.ToInt() {
					continue
				}

				if afterVersion.Minor.ToInt()-1 < 0 {
					return nil, errors.New("unexpected error, out of range of minor versions")
				}

				afterVersion.Minor = data.SubVersionType(strconv.Itoa(afterVersion.Minor.ToInt() - 1))
				if releaseVersion.MajorMinorString() != afterVersion.MajorMinorString() {
					continue
				}
			}

			versions = append(versions, releaseVersion)

			minorVersionCount--
			afterVersion.Minor = data.SubVersionType(strconv.Itoa(afterVersion.Minor.ToInt() - 1))

			if minorVersionCount == 0 {
				break done
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return versions, nil
}

func (g *GithubInfoer) GetLatestVersion(repoOwner string, repo string) (*data.Version, error) {
	release, _, err := g.client.Repositories.GetLatestRelease(context.Background(), repoOwner, repo)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return data.ParseVersion(release.GetTagName()), nil
}
