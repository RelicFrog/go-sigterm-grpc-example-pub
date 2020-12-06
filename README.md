# ARIBOR Project

## Service Architecture

**ARIBOR** is composed of many microservices written in different languages that talk to each other over gRPC.

Find **Protocol Buffers Descriptions** at the [`./pb` directory](./pb).

## New (Core) Service Architecture

* app_frontend
* app_backend
* api_usr_auth
* api_usr_register
* api_usr_config
* api_usr_profile
* api_usr_address
* api_usr_invite
* api_usr_role
* api_usr_group


| Service                                              | Language      | Description                                                                                                                       |
| ---------------------------------------------------- | ------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| [svc_frontend](./src/frontend)                       | NodeJS        | Exposes an HTTP server to serve the website. Does not require signup/login and generates session IDs for all users automatically. |
| [svc_emailer](./src/svc_emailer)                     | Go            | EMail-Service provider, handle all the mail related traffic and request for this application                                      |
| [svc_messenger](./src/svc_messenger)                 | Go            | Messenger-Service stack, handle all the internal messenger functionality for this application                                     |
| [svc_pdfgen](./src/svc_pdfgen)                       | Go            | PDF-/DOC-Generator-Service, handle all the pdf related document creation for this application                                     |
| [svc_usermgr](./src/svc_usermgr)                     | Go            | Primary User-Management-Service, handle registration, profiles and security related tasks for this application                    |
| [svc_loadgen](./src/shippingservice)                 | Python        | Continuously sends requests imitating realistic user shopping flows to the frontend.                                              |

## Installation

We offer the following installation methods:

1. **Running locally** (~20 minutes) You will build
   and deploy microservices images to a single-node Kubernetes cluster running
   on your development machine. There are three options to run a Kubernetes
   cluster locally for this demo:
   - [Minikube](https://github.com/kubernetes/minikube). Recommended for
     Linux hosts (also supports Mac/Windows).
   - [Docker for Desktop](https://www.docker.com/products/docker-desktop).
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
        - [Minikube (recommended for
         Linux)](https://kubernetes.io/docs/setup/minikube/)
        - [Docker for Desktop (recommended for Mac/Windows)](https://www.docker.com/products/docker-desktop)
          - It provides Kubernetes support as [noted
     here](https://docs.docker.com/docker-for-mac/kubernetes/)
        - [Kind](https://github.com/kubernetes-sigs/kind)
   - [skaffold]( https://skaffold.dev/docs/install/) ([ensure version ≥v1.10](https://github.com/GoogleContainerTools/skaffold/releases))
   - Enable GCP APIs for Cloud Monitoring, Tracing, Debugger:
    ```
    gcloud services enable monitoring.googleapis.com \
      cloudtrace.googleapis.com \
      clouddebugger.googleapis.com
    ```
## Load Fixtures

### API-USR-INVITE-CODES
```
docker container kill --signal USR1 aribor-api-grpc-user-invite-code
```

## DeadCode
```
/*
	var validFrom *pb.Date
	var validTo *pb.Date

	validFrom = new(pb.Date)
	validFrom.Day = 1
	validFrom.Month = 1
	validFrom.Year = 2020

	validTo = new(pb.Date)
	validTo.Day = 1
	validTo.Month = 1
	validTo.Year = 2099
 */
```

## Links

- https://itnext.io/learning-go-mongodb-crud-with-grpc-98e425aeaae6
- https://thepracticalsysadmin.com/quickly-securing-local-secrets/
- https://github.com/googleapis/googleapis/blob/master/google/storagetransfer/v1/transfer.proto
- https://github.com/googleapis/googleapis/blob/master/google/type/date.proto
- http://react-material.fusetheme.com/documentation/getting-started/installation
- https://towardsdatascience.com/use-environment-variable-in-your-next-golang-project-39e17c3aaa66
- https://www.mongodb.com/blog/post/mongodb-go-driver-tutorial
- https://github.com/GoogleCloudPlatform/microservices-demo
- https://insujang.github.io/2020-04-04/go-modules/
- https://coolors.co/

### MongoDB Related

- https://www.digitalocean.com/community/tutorials/how-to-use-go-with-mongodb-using-the-mongodb-go-driver-de
- https://kb.objectrocket.com/mongo-db/how-to-create-an-index-using-the-golang-driver-for-mongodb-455
- https://hub.docker.com/r/bitnami/mongodb
- http://learnmongodbthehardway.com/schema/schemabasics/
- https://medium.com/easyread/today-i-learn-text-search-on-mongodb-6b87cd8497c9

### GRPC Related

- https://blog.envoyproxy.io/envoy-and-grpc-web-a-fresh-new-alternative-to-rest-6504ce7eb880
- https://github.com/improbable-eng/grpc-web
- https://github.com/improbable-eng/grpc-web/tree/master/client/grpc-web-react-example
- https://medium.com/swlh/building-a-realtime-dashboard-with-reactjs-go-grpc-and-envoy-7be155dfabfb
- https://github.com/oslabs-beta/ReactRPC

### Envoy Related

- https://github.com/envoyproxy/envoy/issues/7833
- https://medium.com/geckoboard-under-the-hood/we-rolled-out-envoy-at-geckoboard-13c45b4eaddd
- https://github.com/envoyproxy/envoy/tree/master/examples

### UID/UUID Related

- https://www.programmersought.com/article/37021666147/

## Commands
```
go build -mod=readonly -o bin/api_usr_invite $(wildcard *.go)
```

## Commands (mongoDb:shell)
```
db.user_codes.createIndex({meta_code: "text"})
db.user_codes.find({ $text: { $search: "7c42704a-9d87-44ff-9384-79ee3970890e" } });
db.user_codes.find({ meta_valid_to: { $gte: new Date() }})
db.user_codes.getIndexes()
```

## Commands (grpc)
```
curl -v -k --raw -X POST --http2 -H "Content-Type: application/grpc" http://localhost:8080/grpc/user-invite-codes
grpcurl -plaintext 127.0.0.1:50051 aribor.UserInviteCodeService/GetVersion
```

## Commands (node/react)
```
PATH=$PATH:node_modules/.bin
npm install protoc-gen-grpc-web -verbose
protoc -I ../../pb --js_out=import_style=commonjs,binary:./src/proto --grpc-web_out=import_style=commonjs,mode=grpcwebtext:./src/proto ../../pb/rf_example.proto
```

```
import { VersionReq } from '../../../proto/rf_example_pb';
import { UserInviteCodeServicePromiseClient } from '../../../proto/rf_example_grpc_web_pb';

const client = new UserInviteCodeServicePromiseClient('http://localhost:8080', null, null);

		// const stream = client.listInviteCodes(versionRequest, {});

		/*stream.on('data', function (response) {
			setCodes(response.getValue());
		});
		stream.on('status', function(status) {
			console.log(status.code);
			console.log(status.details);
			console.log(status.metadata);
		});
		stream.on('end', function(end) {
			// stream end signal
		});*/
```

```
		/*stream.on('data', function (responseRaw) {
			console.log(responseRaw.toObject());
			this.setState({ data: responseRaw.toObject() });
		});*/
```