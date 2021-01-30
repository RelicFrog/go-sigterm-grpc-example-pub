# gRPC Example Project

## Protocol Architecture

this example project will use many microservices written in different languages that talk to each other over gRPC.

Find **Protocol Buffers Descriptions** at the [`./pb` directory](./pb).

## Service Architecture

| Service                                              | Language      | Description                                                                                                                       |
| ---------------------------------------------------- | ------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| [api_usr_invite](./src/api_usr_invite)               | go            | Provide a gRPC service the handle user invite codes for user registration process                                                 |
| [api_usr_role](./src/api_usr_role)                   | go            | Provide a gRPC service the handle user roles for user management process                                                          |

## Installation

We offer the following installation method:

1. **Running locally** (~20 minutes) You will build
   and deploy microservices images to simple docker compose environment on your development machine:
   
   - [Docker for Desktop / Docker Compose](https://www.docker.com/products/docker-desktop).
    
### Prerequisites

   - [Docker for Desktop (recommended local testing this stack)](https://www.docker.com/products/docker-desktop)
   - [Docker-Compose (for local testing)](https://formulae.brew.sh/formula/docker-compose)
    
### Prepare Service Configuration

   -  Create the ENV files for `sys_mongodb`, the gRPC service `api_user_invite` and `api_user_role` in the appropriate directories for the ENV files currently provided in this example.
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
        ```
        cat <<EOT > src/api_user_role/.env
        DISABLE_TRACING=1
        DISABLE_PROFILER=1
        DISABLE_STATS=0
        DISABLE_DEBUG=0
        DB_MONGO_USR=<your-mongodb-col-usr>
        DB_MONGO_PWD=<your-mongodb-col-pwd>
        DB_MONGO_PDB=<your_mongodb_col>
        DB_MONGO_LNK=mongodb://api_example_user_mongodb:27017/
        PORT=50052
        EOT   
        ```

### Start Service (docker-compose)
```
    docker-compose up -d
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

## Features for next version

   - kubernetes support
   - skaffold/cloudbuild support
   - simple frontend service 

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

Copyright © 2020-2021 Patrick Paechnatz / RelicFrog Team
