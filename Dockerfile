# Build the manager binary
FROM golang:1.16 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
RUN curl -sLo docker-20.10.9.tgz https://download.docker.com/linux/static/stable/x86_64/docker-20.10.9.tgz && \
    tar xzvf docker-20.10.9.tgz
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM ubuntu:latest
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/docker/docker /usr/bin/docker

ENTRYPOINT ["/manager"]
