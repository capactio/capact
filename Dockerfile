FROM golang:1.15.2-alpine as builder

ARG COMPONENT
ARG SOURCE_PATH="./cmd/$COMPONENT/main.go"
ARG BUILD_CMD="go build"

WORKDIR /projectvoltron.dev/voltron

COPY . .
RUN CGO_ENABLED=0 GOARCH=amd64 $BUILD_CMD -ldflags "-s -w" -o /bin/$COMPONENT $SOURCE_PATH

FROM scratch
ARG COMPONENT

COPY --from=builder /bin/$COMPONENT /app

LABEL source=git@github.com:Project-Voltron/go-voltron.git
LABEL app=$COMPONENT

CMD ["/app"]
