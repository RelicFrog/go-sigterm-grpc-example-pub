FROM envoyproxy/envoy:v1.16.0
COPY ./envoy.yaml /etc/envoy/envoy.yaml
CMD /usr/local/bin/envoy -c /etc/envoy/envoy.yaml --component-log-level http:debug,http2:debug,connection:info,grpc:debug
