FROM busybox:latest
RUN mkdir /deploy


COPY ./vespace-provisioner /deploy

# ENV MODULE_VERSION #MODULE_VERSION#

WORKDIR /deploy
CMD ["/deploy/vespace-provisioner"]
