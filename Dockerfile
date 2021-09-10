# syntax = docker/dockerfile:1-experimental

FROM golang:1.16.3-alpine as builder

ARG COMPONENT
ARG SOURCE_PATH="./cmd/$COMPONENT/main.go"
ARG BUILD_CMD="go build"

WORKDIR /capact.io/capact

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

FROM scratch as generic
ARG COMPONENT

# Copy common CA certificates from Builder image (installed by default with ca-certificates package)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /bin/$COMPONENT /app

LABEL source=git@github.com:capactio/capact.git
LABEL app=$COMPONENT

CMD ["/app"]

FROM alpine:3.13.5 as generic-alpine
ARG COMPONENT

# Copy common CA certificates from Builder image (installed by default with ca-certificates package)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /bin/$COMPONENT /app

RUN apk add --no-cache 'git=>2.30' 'openssh=~8.4' && \
    mkdir /root/.ssh && \
    chmod 700 /root/.ssh && \
    ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

LABEL source=git@github.com:capactio/capact.git
LABEL app=$COMPONENT

CMD ["/app"]

FROM builder as e2e-builder

RUN apk add --no-cache 'git=>2.30' && \
    export CGO_ENABLED=0 && go install github.com/onsi/ginkgo/ginkgo@latest

FROM generic as e2e
ARG COMPONENT

COPY --from=builder /bin/$COMPONENT /app.test
COPY --from=e2e-builder /go/bin/ginkgo /ginkgo

CMD ["/ginkgo", "-v", "-nodes=1", "/app.test" ]

FROM alpine:3.13.5 as terraform-runner
ARG COMPONENT

# Copy common CA certificates from Builder image (installed by default with ca-certificates package)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /bin/$COMPONENT /app

RUN apk add --no-cache 'git=>2.30' 'openssh=~8.4' && \
    mkdir /root/.ssh && \
    chmod 700 /root/.ssh && \
    ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

WORKDIR /bin

ENV TERRAFORM_VERSION 1.0.3
RUN wget -nv https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip -O terraform.zip && \
    unzip terraform.zip && \
    rm terraform.zip

WORKDIR /home

LABEL source=git@github.com:capactio/capact.git
LABEL app=$COMPONENT

CMD ["/app"]
