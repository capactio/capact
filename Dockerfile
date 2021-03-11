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
    GOARCH=amd64 $BUILD_CMD -ldflags "-s -w" -o /bin/$COMPONENT $SOURCE_PATH

FROM scratch as generic
ARG COMPONENT

# Copy common CA certificates from Builder image (installed by default with ca-certificates package)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /bin/$COMPONENT /app
COPY hack/mock/graphql /mock/

LABEL source=git@github.com:Project-Voltron/go-voltron.git
LABEL app=$COMPONENT

CMD ["/app"]

FROM alpine:3.12.3 as generic-alpine
ARG COMPONENT

# Copy common CA certificates from Builder image (installed by default with ca-certificates package)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /bin/$COMPONENT /app
COPY test/och-content /test/och-content

RUN apk add --no-cache git=2.26.2-r0 openssh=8.3_p1-r1
RUN mkdir /root/.ssh
RUN chmod 700 /root/.ssh
RUN ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

LABEL source=git@github.com:Project-Voltron/go-voltron.git
LABEL app=$COMPONENT

CMD ["/app"]

FROM alpine:3.12.3  as terraform-runner
ARG COMPONENT

# Copy common CA certificates from Builder image (installed by default with ca-certificates package)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /bin/$COMPONENT /app

ENV TERRAFORM_VERSION 0.14.6

WORKDIR /bin
RUN \
    wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip -O terraform.zip && \
    unzip terraform.zip && rm terraform.zip

COPY hack/runners/terraform /workspace

WORKDIR /workspace
RUN /bin/terraform init
RUN rm /workspace/providers.tf

LABEL source=git@github.com:Project-Voltron/go-voltron.git
LABEL app=$COMPONENT

CMD ["/app"]
