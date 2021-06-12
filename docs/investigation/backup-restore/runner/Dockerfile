FROM golang:1.16.2-alpine

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-s -w" -o /runner runner.go

RUN wget https://github.com/vmware-tanzu/velero/releases/download/v1.5.3/velero-v1.5.3-linux-amd64.tar.gz && \
    tar -xvf velero-v1.5.3-linux-amd64.tar.gz && cp velero-v1.5.3-linux-amd64/velero /velero

CMD ["/runner"]

