#!/usr/bin/env bash
# ref: https://ignite.readthedocs.io/en/stable/installation/

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

GOARCH=$(go env GOARCH 2>/dev/null || echo "amd64")

KUBEFIRE_VERSION=${KUBEFIRE_VERSION:-}
CONTAINERD_VERSION=${CONTAINERD_VERSION:-""}
IGNITE_VERSION=${IGNITE_VERSION:-""}
TAILSCALE_VERSION=${TAILSCALE_VERSION:-""}
CNI_VERSION=${CNI_VERSION:-""}
RUNC_VERSION=${RUNC_VERSION:-""}

if [ -z "$KUBEFIRE_VERSION" ] || [ -z "$CONTAINERD_VERSION" ] || [ -z "$IGNITE_VERSION" ] || [ -z "$TAILSCALE_VERSION" ] || [ -z "$CNI_VERSION" ] || [ -z "$RUNC_VERSION" ]; then
  echo "incorrect versions provided!" >/dev/stderr
  exit 1
fi

STABLE_KUBEFIRE_VERSION=$(sed -E "s/(v[0-9]+\.[0-9]+\.[0-9]+)[a-zA-Z0-9\-]*/\1/g"< <(echo "$KUBEFIRE_VERSION"))

DOWNLOAD_DIR=/home/marc/projects/NEW/edgi-backend/edgi-pulumi/kubefire
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
TMP_DIR=$( mktemp -d -p /dev/shm )
pushd $TMP_DIR

function cleanup() {
  rm -rf $TMP_DIR || true
}

trap cleanup EXIT ERR INT TERM

function _check_version() {
  set +o pipefail

  local exec_name=$1
  local exec_version_cmd=$2
  local version=$3

  command -v "${exec_name}" && [[ "$(eval "$exec_name $exec_version_cmd 2>&1")" =~ $version ]]
  return $?
}

function _is_arm_arch() {
    uname -m | grep "aarch64"
    return $?
}

function check_virtualization() {
  if _is_arm_arch; then
    return
  fi

  lscpu | grep "Virtuali[s|z]ation"
  lsmod | grep kvm
}

function install_tailscale() {
  if _check_version /usr/local/bin/tailscale version $TAILSCALE_VERSION; then
    echo "tailscale (${TAILSCALE_VERSION}) installed already!"
    return
  fi

  "$SCRIPT_DIR"/docker-image-extract.sh tailscale/tailscale:latest "$TMP_DIR"/tailscale-docker
  sudo cp "$TMP_DIR"/tailscale-docker/usr/local/bin/tailscale /usr/local/bin
  sudo cp "$TMP_DIR"/tailscale-docker/usr/local/bin/tailscaled /usr/local/bin
  rm -rf "$TMP_DIR"/tailscale-docker

  curl -sfSLO "https://raw.githubusercontent.com/tailscale/tailscale/v${TAILSCALE_VERSION}/cmd/tailscaled/tailscaled.service"
  sudo groupadd tailscaled || true
  sudo mv tailscaled.service /etc/systemd/system/tailscaled.service
  chgrp_path=$(command -v chgrp | tr -d '\n')
  sudo sed -i -E "s#/usr/sbin/#/usr/local/bin/#g" /etc/systemd/system/tailscaled.service
  sudo sed -i -E "s#(ExecStart=/usr/local/bin/tailscaled.*)#\1\nExecStartPost=${chgrp_path} tailscaled /var/lib/tailscale/tailscaled.state /run/tailscale/tailscaled.sock#g" /etc/systemd/system/tailscaled.service

  sudo systemctl enable --now tailscaled
}

function check_tailscale() {
  tailscale version
}

function install_containerd() {
  if _check_version /usr/local/bin/containerd --version $CONTAINERD_VERSION; then
    echo "containerd (${CONTAINERD_VERSION}) installed already!"
    return
  fi

  local version="${CONTAINERD_VERSION:1}"
  local dir=containerd-$version


  if _is_arm_arch; then
    echo "!!! Please install containerd aarch64 via system package manager, because there is no official aarch64 release from the github repo. !!!"
    return
  fi

  curl -sfSLO "https://github.com/containerd/containerd/releases/download/${CONTAINERD_VERSION}/containerd-${version}-linux-${GOARCH}.tar.gz"
  mkdir -p $dir
  tar -zxvf $dir*.tar.gz -C $dir
  chmod +x $dir/bin/*
  sudo mv $dir/bin/* /usr/local/bin/

  curl -sfSLO "https://raw.githubusercontent.com/containerd/containerd/${CONTAINERD_VERSION}/containerd.service"
  sudo groupadd containerd || true
  sudo mv containerd.service /etc/systemd/system/containerd.service

  chgrp_path=$(command -v chgrp | tr -d '\n')
  sudo sed -i -E "s#(ExecStart=/usr/local/bin/containerd)#\1\nExecStartPost=${chgrp_path} containerd /run/containerd/containerd.sock#g" /etc/systemd/system/containerd.service

  sudo mkdir -p /etc/containerd
  containerd config default | sudo tee /etc/containerd/config.toml >/dev/null
  sudo systemctl enable --now containerd
}

function install_runc() {
  if _check_version /usr/local/bin/runc -version $RUNC_VERSION; then
    echo "runc (${RUNC_VERSION}) installed already!"
    return
  fi

  if _is_arm_arch; then
    echo "!!! Please install runc aarch64 via system package manager, because there is no official aarch64 release from the github repo. !!!"
    return
  fi

  curl -sfSL "https://github.com/opencontainers/runc/releases/download/${RUNC_VERSION}/runc.amd64" -o runc
  chmod +x runc
  sudo mv runc /usr/local/bin/
}

function install_cni() {
  if _check_version /opt/cni/bin/bridge --version $CNI_VERSION; then
    echo "CNI plugins (${CNI_VERSION}) installed already!"
    return
  fi

  mkdir -p /opt/cni/bin

  local f="https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-amd64-${CNI_VERSION}.tgz"
  if _is_arm_arch; then
    f="https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-arm64-${CNI_VERSION}.tgz"
  fi

  curl -sfSL "$f" | tar -C /opt/cni/bin -xz
}

function install_cni_patches() {
    if _is_arm_arch; then
      #curl -o host-local-rev -sSL "https://github.com/edgi-io/kubefire/releases/download/${STABLE_KUBEFIRE_VERSION}/host-local-rev-linux-arm64" # FIXME: add -f back later

      #HACK
      cp "$DOWNLOAD_DIR"/target/cni/host-local-rev-linux-arm64 host-local-rev
    else
      #curl -o host-local-rev -sfSL "https://github.com/edgi-io/kubefire/releases/download/${STABLE_KUBEFIRE_VERSION}/host-local-rev-linux-amd64" || \
      #curl -o host-local-rev -sfSL "https://github.com/edgi-io/kubefire/releases/download/${STABLE_KUBEFIRE_VERSION}/host-local-rev"

      #HACK
      cp "$DOWNLOAD_DIR"/target/cni/host-local-rev-linux-amd64 host-local-rev
    fi

    chmod +x host-local-rev
    sudo mv host-local-rev /opt/cni/bin/
}

function install_ignite() {
  if _check_version /usr/local/bin/ignite version $IGNITE_VERSION; then
    echo "ignite (${IGNITE_VERSION}) installed already!"
    return
  fi

  for binary in ignite ignited; do
    echo "Installing $binary..."

    local f="https://github.com/weaveworks/ignite/releases/download/${IGNITE_VERSION}/${binary}-amd64"
    if _is_arm_arch; then
      f="https://github.com/weaveworks/ignite/releases/download/${IGNITE_VERSION}/${binary}-arm64"
    fi

    curl -sfSLo $binary "$f"
    chmod +x $binary
    sudo mv $binary /usr/local/bin
  done
}

function check_ignite() {
  ignite version
}

function create_cni_default_config() {
  mkdir -p /etc/cni/net.d/ || true
  sudo cat <<'EOF' > /etc/cni/net.d/00-kubefire.conflist
{
	"cniVersion": "0.4.0",
	"name": "kubefire-cni-bridge",
	"plugins": [
		{
			"type": "bridge",
			"bridge": "kubefire0",
			"isGateway": true,
			"isDefaultGateway": true,
			"promiscMode": true,
			"ipMasq": true,
			"ipam": {
				"type": "host-local-rev",
				"subnet": "10.62.0.0/16"
			}
		},
		{
			"type": "portmap",
			"capabilities": {
				"portMappings": true
			}
		},
		{
			"type": "firewall"
		}
	]
}
EOF
}

check_virtualization

install_runc
install_containerd
install_cni
install_cni_patches
install_ignite
check_ignite
create_cni_default_config
install_tailscale
check_tailscale

popd
