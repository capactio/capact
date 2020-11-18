# syntax = docker/dockerfile:1-experimental

FROM golang:1.15.2-alpine as builder

ARG COMPONENT
ARG SOURCE_PATH="./cmd/$COMPONENT/main.go"
ARG BUILD_CMD="go build"

WORKDIR /projectvoltron.dev/voltron

# Use experimental frontend syntax to cache dependencies.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Use experimental frontend syntax to cache go build.
# Replace `COPY . .` with `--mount=target=.` to speed up as we do not need them to persist in the final image.
# https://github.com/moby/buildkit/blob/master/frontend/dockerfile/docs/experimental.md
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOARCH=amd64 $BUILD_CMD -ldflags "-s -w" -o /bin/$COMPONENT $SOURCE_PATH

FROM scratch
ARG COMPONENT

COPY --from=builder /bin/$COMPONENT /app
COPY hack/mock/$COMPONENT/ /mock/$COMPONENT

LABEL source=git@github.com:Project-Voltron/go-voltron.git
LABEL app=$COMPONENT

CMD ["/app"]
