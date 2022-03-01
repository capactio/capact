# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

WORKDIR /app

COPY . ./
RUN go mod download

RUN go build -o /lambda

EXPOSE 8686

CMD [ "/lambda" ]
