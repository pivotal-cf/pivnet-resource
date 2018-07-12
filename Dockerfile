FROM golang:alpine as builder
COPY . /go/src/github.com/pivotal-cf/pivnet-resource
ENV CGO_ENABLED 0
RUN go build -o /assets/in github.com/pivotal-cf/pivnet-resource/cmd/in
RUN go build -o /assets/out github.com/pivotal-cf/pivnet-resource/cmd/out
RUN go build -o /assets/check github.com/pivotal-cf/pivnet-resource/cmd/check

FROM alpine:edge AS resource
RUN apk add --no-cache bash tzdata ca-certificates unzip zip gzip tar
COPY --from=builder assets/ /opt/resource/
RUN chmod +x /opt/resource/*
