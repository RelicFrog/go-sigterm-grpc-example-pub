# Copyright 2020 Team RelicFrog
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

#
# builder image
# --
FROM golang:1.15-alpine AS builder
RUN apk add --no-cache ca-certificates git
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /service/api-usr-invite-codes .

#
# release image
# --
FROM alpine AS release
RUN apk add --no-cache ca-certificates
RUN GRPC_HEALTH_PROBE_VERSION=v0.2.0 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe
WORKDIR /service/api-usr-invite-codes
COPY --from=builder /service/api-usr-invite-codes ./server
COPY .env /service/api-usr-invite-codes
EXPOSE 50051
ENTRYPOINT ["/service/api-usr-invite-codes/server"]
