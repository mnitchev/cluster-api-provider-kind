# syntax = docker/dockerfile:experimental
FROM golang:1.17 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY infrastructure/ infrastructure/
COPY k8s/ k8s/
RUN curl -sLo docker-20.10.9.tgz https://download.docker.com/linux/static/stable/x86_64/docker-20.10.9.tgz && \
    tar xzvf docker-20.10.9.tgz

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o manager main.go

FROM gcr.io/distroless/static:latest
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/docker/docker /usr/bin/docker

ENTRYPOINT ["/manager"]
