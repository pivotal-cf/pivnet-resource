FROM pivnet/golang

ADD cmd/check/check /opt/resource/check
ADD cmd/in/in /opt/resource/in
ADD cmd/out/out /opt/resource/out

RUN chmod +x /opt/resource/*