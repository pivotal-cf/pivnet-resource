FROM gliderlabs/alpine:3.2

RUN apk --update add \
  ca-certificates \
  jq

ADD check /opt/resource/check
ADD in /opt/resource/in
ADD out /opt/resource/out
ADD s3-out /opt/resource/s3-out

RUN chmod +x /opt/resource/*
