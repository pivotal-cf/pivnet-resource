FROM concourse/buildroot:git

RUN \
  cd /usr/bin/ && \
  curl \
    -L \
    -O \
    https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 && \
  mv jq-linux64 jq && \
  chmod +x jq

ADD cmd/check/check /opt/resource/check
ADD cmd/in/in /opt/resource/in
ADD cmd/out/out /opt/resource/out
ADD s3-out /opt/resource/s3-out

RUN chmod +x /opt/resource/*
