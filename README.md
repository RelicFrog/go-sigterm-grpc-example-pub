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
| [app_frontend](./src/app_frontend)                   | Go            | Exposes an HTTP server to serve the frontend website.                                                                             |
| [app_backend](./src/app_backend)                     | Go            | Exposes an HTTP server to serve the backend website.                                                                              |
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

## Load Fixtures

### user invite code gRPC API `api_usr_invite`
```
docker container kill --signal USR1 rf-example-grpc-user-invite-code
```
