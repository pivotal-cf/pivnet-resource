FROM gliderlabs/alpine:3.2

RUN apk --update add \
  ca-certificates \
  jq

ADD check /opt/resource/check
ADD in /opt/resource/in
ADD out /opt/resource/out
RUN chmod +x /opt/resource/*
