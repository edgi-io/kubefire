package di

import (
	"github.com/edgi-io/kubefire/pkg/bootstrap"
	"github.com/edgi-io/kubefire/pkg/bootstrap/versionfinder"
	"github.com/edgi-io/kubefire/pkg/cache"
	"github.com/edgi-io/kubefire/pkg/cluster"
	pkgconfig "github.com/edgi-io/kubefire/pkg/config"
	"github.com/edgi-io/kubefire/pkg/node"
	"github.com/edgi-io/kubefire/pkg/output"
)

type BootstrapperAware interface {
	SetBootstrapper(bootstrapper bootstrap.Bootstrapper)
}

type VersionFinderAware interface {
	SetVersionFinder(versionFinder versionfinder.Finder)
}

type ConfigManagerAware interface {
	SetConfigManager(configManager pkgconfig.Manager)
}

type ClusterManagerAware interface {
	SetClusterManager(clusterManager cluster.Manager)
}

type NodeManagerAware interface {
	SetNodeManager(nodeManager node.Manager)
}

type OutputAware interface {
	SetOutputer(outputer output.Outputer)
}

type CacheManagerAware interface {
	SetCacheManager(cacheManager cache.Manager)
}
