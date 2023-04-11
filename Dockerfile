FROM ubuntu:jammy

RUN apt-get update && apt-get install -y unzip ca-certificates && rm -rf /var/lib/apt/lists/*

COPY cmd/check/check /opt/resource/check
COPY cmd/in/in /opt/resource/in
COPY cmd/out/out /opt/resource/out
