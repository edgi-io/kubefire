#!/bin/sh
#
# This script pulls and extracts all files from an image in Docker Hub.
#
# Copyright (c) 2020-2021, Jeremy Lin
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
# FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
# DEALINGS IN THE SOFTWARE.
#
if [ $# -ne 2 ]; then
    exit 0
    echo "$0 IMAGE[:REF] DEST"
    echo
    echo "This script pulls and extracts all files from an image in Docker Hub."
    echo
    echo "Examples:"
    echo
    echo "# Pull and extract all files in the 'hello-world' image tagged 'latest'."
    echo "\$ $0 hello-world:latest ./output"
    echo
    echo "# Same as above; tag defaults to 'latest'."
    echo "\$ $0 hello-world ./output"
    echo
    echo "# Same as above, but specify the image by digest."
    echo "# This also allows for pulling an image for a non-amd64 platform."
    echo "\$ $0 hello-world:sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042 ./output"
    exit 1
fi

have_curl() {
    command -v curl >/dev/null
}

have_wget() {
    command -v wget >/dev/null
}

if ! have_curl && ! have_wget; then
    echo "This script requires either curl or wget."
    exit 1
fi

image_spec="$1"
image="${image_spec%%:*}"
if [ "${image#*/}" = "${image}" ]; then
    # Docker official images are in the 'library' namespace.
    image="library/${image}"
fi
tag="${image_spec#*:}"
if [ "${tag}" = "${image_spec}" ]; then
    tag=latest
fi

# Given a JSON input on stdin, extract the string value associated with the
# specified key. This avoids an extra dependency on a tool like `jq`.
extract() {
    local key="$1"
    # Extract "<key>":"<val>" (assumes key/val won't contain double quotes).
    # The colon may have whitespace on either side.
    grep -o "\"${key}\"[[:space:]]*:[[:space:]]*\"[^\"]\+\"" |
    # Extract just <val> by deleting the last '"', and then greedily deleting
    # everything up to '"'.
    sed -e 's/"$//' -e 's/.*"//'
}

# Fetch a URL to stdout. Up to two header arguments may be specified:
#
#   fetch <url> [name1: value1] [name2: value2]
#
fetch() {
    if have_curl; then
        if [ $# -eq 2 ]; then
            set -- -H "$2" "$1"
        elif [ $# -eq 3 ]; then
            set -- -H "$2" -H "$3" "$1"
        fi
        curl -sSL "$@"
    else
        if [ $# -eq 2 ]; then
            set -- --header "$2" "$1"
        elif [ $# -eq 3 ]; then
            set -- --header "$2" --header "$3" "$1"
        fi
        wget -qO- "$@"
    fi
}

# https://docs.docker.com/registry/spec/auth/token/#how-to-authenticate
api_token_url="https://auth.docker.io/token?service=registry.docker.io&scope=repository:$image:pull"

# https://github.com/docker/distribution/blob/master/docs/spec/api.md#pulling-an-image-manifest
manifest_url="https://registry-1.docker.io/v2/${image}/manifests/$tag"

# https://github.com/docker/distribution/blob/master/docs/spec/api.md#pulling-a-layer
blobs_base_url="https://registry-1.docker.io/v2/${image}/blobs"

echo "Getting API token..."
token=$(fetch "${api_token_url}" | extract 'token')
auth_header="Authorization: Bearer $token"
v2_header="Accept: application/vnd.docker.distribution.manifest.v2+json"

echo "Getting image manifest for $image:$tag..."
layers=$(fetch "${manifest_url}" "${auth_header}" "${v2_header}" |
             # Extract `digest` values only after the `layers` section appears.
             sed -n '/"layers":/,$ p' |
             extract 'digest')

if [ -z "${layers}" ]; then
    echo "No layers returned. Verify that the image and tag are valid."
    exit 1
fi

output_dir="$2"
mkdir -p "${output_dir}"

for layer in $layers; do
    hash="${layer#sha256:}"
    echo "Fetching and extracting layer ${hash}..."
    fetch "${blobs_base_url}/${layer}" "${auth_header}" | gzip -d | tar -C "${output_dir}" -xf -
    # Ref: https://github.com/moby/moby/blob/master/image/spec/v1.2.md#creating-an-image-filesystem-changeset
    #      https://github.com/moby/moby/blob/master/pkg/archive/whiteouts.go
    # Search for "whiteout" files to indicate files deleted in this layer.
    OLD_IFS="${IFS}"
    find "${output_dir}" -name '.wh.*' | while IFS= read -r f; do
        dir="${f%/*}"
        wh_file="${f##*/}"
        file="${wh_file#.wh.}"
        # Delete both the whiteout file and the whited-out file.
        rm -rf "${dir}/${wh_file}" "${dir}/${file}"
    done
    IFS="${OLD_IFS}"
done

echo "Image contents extracted into $output_dir."
