FROM gliderlabs/alpine:3.2

RUN apk --update add \
  ca-certificates \
  jq

ADD check /opt/resource/check
ADD in /opt/resource/in
ADD out /opt/resource/out

copy ./scripts/test-all-the-things /test-all-the-things

ENTRYPOINT "/test-all-the-things"
