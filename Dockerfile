FROM progrium/busybox
MAINTAINER Peter Collins <peter@drifty.com>

ADD ./stage/cryptographer /bin/cryptographer

ENV DOCKER_HOST unix:///tmp/docker.sock

ENTRYPOINT ["/bin/cryptographer"]
