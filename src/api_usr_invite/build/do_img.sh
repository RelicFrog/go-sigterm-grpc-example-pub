#!/bin/bash -eu
#
# Copyright 2020-2021 Team RelicFrog
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SCRIPT_PATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
SCRIPT_NAME="$(basename "$(test -L "$0" && readlink "$0" || echo "$0")")"

metaServiceName="user-manager"
metaServiceCore="relicFrog"
metaImageName="rf-example-grpc-user-invite-code"
metaImageVersion="1.0.0"

get_password () {
  pass show "$1"
}

#
# entrypoint
# --

echo -e "script [$metaServiceCore/$metaServiceName/$SCRIPT_NAME] build service image\n"

cd ${SCRIPT_PATH}
docker build --tag ${metaImageName}:${metaImageVersion} ../. ; cd -
