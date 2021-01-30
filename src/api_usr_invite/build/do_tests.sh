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
metaServiceCore="grpc-example"
metaMongoDbVersion="4.4.1"
metaMongoDbLocalTargetPort="27018"
metaMongoDbLocalTargetSvcName="usermgr-mongo-db-test"
metaMongoDbTestRootPwd="dummyrootpwd123"
metaMongoDbTestDbName="example_user_db"
metaMongoDbTestUser="test_db_col_usr"
metaMongoDbTestPwd="test_db_col_pwd123"
metaRuntimeHash=$(openssl rand -hex 4)

tear_up () {

  if [ ! "$(docker ps -q -f name=${metaMongoDbLocalTargetSvcName}-${metaRuntimeHash})" ]; then
    if [ "$(docker ps -aq -f status=exited -f name=${metaMongoDbLocalTargetSvcName}-${metaRuntimeHash})" ]; then
        echo -e "[$metaServiceCore/$metaServiceName/$SCRIPT_NAME] remove old mongodb container (testing scope) <remove>\n"
        docker rm -fv ${metaMongoDbLocalTargetSvcName}-${metaRuntimeHash}
    fi

    echo -e "script [$metaServiceCore/$metaServiceName/$SCRIPT_NAME] init new mongodb container (testing scope) <create>\n"
    docker run -d --name ${metaMongoDbLocalTargetSvcName}-${metaRuntimeHash} -p ${metaMongoDbLocalTargetPort}:27017 \
        -e MONGODB_ROOT_PASSWORD=${metaMongoDbTestRootPwd} \
        -e MONGODB_USERNAME=${metaMongoDbTestUser} \
        -e MONGODB_PASSWORD=${metaMongoDbTestPwd} \
        -e MONGODB_DATABASE=${metaMongoDbTestDbName} \
        bitnami/mongodb:${metaMongoDbVersion} >/dev/null 2>&1
  else
      echo -e "script [$metaServiceCore/$metaServiceName/$SCRIPT_NAME] mongodb container (testing scope) is running <ignore>\n"
  fi
}

tear_down () {

  if [ "$(docker ps -q -f name=${metaMongoDbLocalTargetSvcName}-${metaRuntimeHash})" ]; then
    echo -e "\nscript [$metaServiceCore/$metaServiceName/$SCRIPT_NAME] remove mongodb container (testing scope) <remove>\n"
    docker rm -fv ${metaMongoDbLocalTargetSvcName}-${metaRuntimeHash} >/dev/null 2>&1
  fi

}

run_tests () {

  tear_up

  cd ${SCRIPT_PATH}
  GOPATH=$HOME/go GOBIN=$HOME/go/bin DB_MONGO_USR=${metaMongoDbTestUser} DB_MONGO_PWD=${metaMongoDbTestPwd} DB_MONGO_LNK=mongodb://localhost:${metaMongoDbLocalTargetPort} DISABLE_DEBUG=1 go test ../ -mod=readonly -v
  cd -

  tear_down
}

#
# -- entrypoint
#

run_tests