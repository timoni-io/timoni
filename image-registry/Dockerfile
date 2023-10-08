FROM scratch
ARG TARGETPLATFORM

COPY _build/bin/$TARGETPLATFORM/image-registry /image-registry

WORKDIR /
CMD ["/image-registry"]
