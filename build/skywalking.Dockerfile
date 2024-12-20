# syntax=docker.io/docker/dockerfile:1.10
# https://docs.docker.com/build/dockerfile/frontend/

FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.23-alpine AS build

ARG MODULE_NAME
ARG APP_NAME
ARG VERSION
ARG BUILDTIME

ADD https://github.com/apache/skywalking-go.git#v0.5.0 /skywalking-go
RUN --mount=type=cache,target=/root/.cache/go-build/,id=build-sw-$TARGETARCH \
    --mount=type=cache,target=/go/pkg/mod/ \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -C /skywalking-go/tools/go-agent -v -ldflags " \
    -X 'main.version=0.5.0' \
    -X 'main.goVersion=$(go env GOVERSION)' \
    -X 'main.gitCommit=$(git -C /skywalking-go rev-parse --short HEAD)'" \
    -o /usr/local/bin/skywalking-go-agent ./cmd

# https://docs.docker.com/build/guide/mounts/
WORKDIR /app
RUN --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=cache,target=/go/pkg/mod/ \
    go mod download -x

# https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build/,id=build-sw-$TARGETARCH \
    --mount=type=cache,target=/go/pkg/mod/ \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -toolexec="skywalking-go-agent" -tags=skywalking \
    -trimpath -ldflags="-s -w \
    -X '${MODULE_NAME}/global.ModName=${MODULE_NAME}' \
    -X '${MODULE_NAME}/global.AppName=${APP_NAME}-sw' \
    -X '${MODULE_NAME}/global.Version=${VERSION}' \
    -X '${MODULE_NAME}/global.BuildTime=${BUILDTIME}'" \
    -o main .

FROM gcr.io/distroless/static-debian12:latest
COPY --from=build /app/main /main
ENTRYPOINT ["/main"]
