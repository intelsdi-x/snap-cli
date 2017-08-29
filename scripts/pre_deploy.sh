#!/usr/bin/env bash
# File managed by pluginsync

# http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
# Copyright 2016 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

build_path="${__proj_dir}/build"
_info "build_path: ${build_path}"
_debug "$(find "${build_path}")"

plugin_name="${__proj_dir##*/}"
git_sha=$(git log --pretty=format:"%H" -1)
s3_path="${__proj_dir}/s3/${plugin_name}"

set +u
if [ -z "$TRAVIS_TAG" ]; then
  set -u
  git_path="${s3_path}/${git_sha}"
  latest_path="${s3_path}/latest_build"
  mkdir -p "${git_path}"
  mkdir -p "${latest_path}"

  _info "copying snap cli binary to ${git_path}"
  cp -rp "${build_path}/"* "${git_path}"
  _info "copying snap cli binary to ${latest_path}"
  cp -rp "${build_path}/"* "${latest_path}"
else
  set -u
  tag_path="${s3_path}/${TRAVIS_TAG}"
  latest_path="${s3_path}/latest"
  mkdir -p "${tag_path}"
  mkdir -p "${latest_path}"

  _info "copying snap cli binary to ${tag_path}"
  cp -rp "${build_path}/"* "${tag_path}"
  _info "copying snap cli binary to ${latest_path}"
  cp -rp "${build_path}/"* "${latest_path}"
fi

release_path="${SNAP_PATH:-"${__proj_dir}/release"}"
mkdir -p "${release_path}"

_info "packaging snap cli binary to ${release_path}"

filename=snaptel

for file in "${build_path}"/**/* ; do
  arch="${file##*/}"
  parent="${file%/*}"
  os="${parent##*/}"

  tar -cvzf "${release_path}/${filename}_${os}_${arch}.tar.gz" -C "${file}" . > /dev/null

done

_debug "$(find "${build_path}")"
_debug "$(find "${s3_path}")"
_debug "$(find "${release_path}")"
