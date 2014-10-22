FROM progrium/busybox
MAINTAINER Peter Collins <peter@drifty.com>

ADD ./stage/cryptographer /bin/cryptographer

ENV SECRETS_DIR /secrets
ENV KEY_RING /var/usr/keyring.gpg
ENV DOCKER_HOST unix:///tmp/docker.sock

VOLUME ["/secrets"]

ENTRYPOINT ["/bin/cryptographer"]
