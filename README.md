# gRPC Example Project

## Protocol Architecture

this example project will use many microservices written in different languages that talk to each other over gRPC.

Find **Protocol Buffers Descriptions** at the [`./pb` directory](./pb).

## Service Architecture

* app_frontend
* app_backend
* api_usr_invite

| Service                                              | Language      | Description                                                                                                                       |
| ---------------------------------------------------- | ------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| [app_frontend](./src/app_frontend)                   | Go            | Exposes an HTTP server to serve the frontend website (_not implemented yet_).                                                     |
| [app_backend](./src/app_backend)                     | Go            | Exposes an HTTP server to serve the backend website (_not implemented yet_).                                                      |
| [api_usr_invite](./src/api_usr_invite)               | Go            | Provide a gRPC service the handle user invite codes for our upcoming user registration process                                    |

## Installation

We offer the following installation methods:

1. **Running locally** (~20 minutes) You will build
   and deploy microservices images to a single-node Kubernetes cluster running
   on your development machine. There are three options to run a Kubernetes
   cluster locally for this demo:
   - [Minikube](https://github.com/kubernetes/minikube). Recommended for
     Linux hosts (also supports Mac/Windows).
   - [Docker for Desktop / Docker Compose](https://www.docker.com/products/docker-desktop).
     Recommended for Mac/Windows.
   - [Kind](https://kind.sigs.k8s.io). Supports Mac/Windows/Linux.

1. **Running on Google Kubernetes Engine (GKE)** (~30 minutes) You will build,
   upload and deploy the container images to a Kubernetes cluster on Google
   Cloud.

1. **Using pre-built container images:** (~10 minutes, you will still need to
   follow one of the steps above up until `skaffold run` command). With this
   option, you will use pre-built container images that are available publicly,
   instead of building them yourself, which takes a long time).

### Prerequisites

   - kubectl (can be installed via `gcloud components install kubectl`)
   - Local Kubernetes cluster deployment tool:
        - [Minikube (recommended for Linux)](https://kubernetes.io/docs/setup/minikube/)
        - [Docker for Desktop (recommended for Mac/Windows)](https://www.docker.com/products/docker-desktop)
            - It provides Kubernetes support as [noted here](https://docs.docker.com/docker-for-mac/kubernetes/)
        - [Kind](https://github.com/kubernetes-sigs/kind)
   - [skaffold]( https://skaffold.dev/docs/install/) ([ensure version ≥v1.10](https://github.com/GoogleContainerTools/skaffold/releases))
   - Enable GCP APIs for Cloud Monitoring, Tracing, Debugger:
    ```
    gcloud services enable monitoring.googleapis.com \
      cloudtrace.googleapis.com \
      clouddebugger.googleapis.com
    ```
    
### Prepare Service Configuration

1. Create the ENV files for `sys_mongodb`, and the gRPC service `api_user_invite` in the appropriate directories for the ENV files currently provided in this example.
    ```
    cat <<EOT > src/sys_mongodb/.env
    MONGODB_ROOT_PASSWORD=<your-mongodb-root-pwd>
    MONGODB_USERNAME=<your-mongodb-col-usr>
    MONGODB_PASSWORD=<your-mongodb-col-pwd>
    MONGODB_DATABASE=<your_mongodb_col>
    EOT   
    ```
    ```
    cat <<EOT > src/api_user_invite/.env
    DISABLE_TRACING=1
    DISABLE_PROFILER=1
    DISABLE_STATS=0
    DISABLE_DEBUG=0
    DB_MONGO_USR=<your-mongodb-col-usr>
    DB_MONGO_PWD=<your-mongodb-col-pwd>
    DB_MONGO_PDB=<your_mongodb_col>
    DB_MONGO_LNK=mongodb://api_example_user_mongodb:27017/
    PORT=50051
    EOT   
    ```
2. Create the gRPC service image file for `api_user_invite`
    ```
    cd src/api_user_invite
    make image
    ```

### Start Service (docker-compose)
```
    docker-compose up
```

### Test ICP Signal Scope (in docker-compose / native docker)

1. loading fixtures / seeding database
    ```
    docker container kill --signal USR1 rf-example-grpc-user-invite-code
    ```
2. activate / deactivate random request latency effect
    ```
    docker container kill --signal USR2 rf-example-grpc-user-invite-code
    ```

### Test ICP Signal Scope (in kubernetes)

1. loading fixtures / seeding database
    ```
    kubectl exec \
        $(kubectl get pods -l app=rf-example-grpc-user-invite-code -o jsonpath='{.items[0].metadata.name}') \
        -c server -- kill -USR1 1   
    ```
2. activate / deactivate random request latency effect
    ```
    kubectl exec \
        $(kubectl get pods -l app=rf-example-grpc-user-invite-code -o jsonpath='{.items[0].metadata.name}') \
        -c server -- kill -USR1 2
    ```

## License

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) 

See [LICENSE](LICENSE) for full details.

    Licensed to the Apache Software Foundation (ASF) under one
    or more contributor license agreements.  See the NOTICE file
    distributed with this work for additional information
    regarding copyright ownership.  The ASF licenses this file
    to you under the Apache License, Version 2.0 (the
    "License"); you may not use this file except in compliance
    with the License.  You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing,
    software distributed under the License is distributed on an
    "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
    KIND, either express or implied.  See the License for the
    specific language governing permissions and limitations
    under the License.

## Copyright

Copyright © 2019-2020 Patrick Paechnatz / RelicFrog Team
