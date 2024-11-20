FROM --platform=${BUILDPLATFORM} golang:1.23 as build-env

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apt-get install -yq --no-install-recommends git

# Copy source + vendor
COPY . /go/src/github.com/vincejv/grafana-matrix-forwarder
WORKDIR /go/src/github.com/vincejv/grafana-matrix-forwarder

# Compile go binaries
ENV GOPATH=/go
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GO111MODULE=on go build -v -a -ldflags "-s -w" -o /go/bin/grafana-matrix-forwarder .

# Build final image from alpine
FROM --platform=${TARGETPLATFORM} alpine:latest
RUN apk --update --no-cache add curl && rm -rf /var/cache/apk/*
COPY --from=build-env /go/bin/grafana-matrix-forwarder /usr/bin/grafana-matrix-forwarder

# Create a group and user
RUN addgroup -S grafana-matrix-forwarder && adduser -S grafana-matrix-forwarder -G grafana-matrix-forwarder
USER grafana-matrix-forwarder

ENTRYPOINT ["grafana-matrix-forwarder"]

EXPOSE 8092