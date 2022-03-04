package node

import (
	"fmt"
	"github.com/edgi-io/kubefire/pkg/config"
	"github.com/edgi-io/kubefire/pkg/data"
	"regexp"
	"strconv"
	"time"
)

type Type string

const (
	Master Type = "master"
	Worker Type = "worker"

	//NameFormat node name format: <cluster name>-<node type>-<node index>
	NameFormat = "%s-%s-%s"
)

var namePattern = fmt.Sprintf(`%%s-(%s|%s)-\d+`, Master, Worker)

type Manager interface {
	CreateNodes(nodeType Type, node *config.Node, started bool) error
	DeleteNodes(nodeType Type, node *config.Node) error
	DeleteNode(name string) error
	GetNode(name string) (*data.Node, error)
	ListNodes(clusterName string) ([]*data.Node, error)
	LoginBySSH(name string, configManager config.Manager) error
	WaitNodesRunning(clusterName string, timeoutMin time.Duration) error
	StartNodes(clusterName string) error
	StartNode(name string) error
	StopNodes(clusterName string) error
	StopNode(name string) error
	GetCaches() ([]interface{}, error)
	DeleteCaches() error
}

func Name(clusterName string, nodeType Type, index int) string {
	return fmt.Sprintf(NameFormat, clusterName, nodeType, strconv.Itoa(index))
}

func IsValidNodeName(nodeName string, clusterName string) bool {
	re, _ := regexp.Compile(fmt.Sprintf(namePattern, clusterName))
	return re.MatchString(nodeName)
}
