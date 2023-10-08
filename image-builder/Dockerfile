FROM --platform=$TARGETPLATFORM docker:20.10-dind
ARG TARGETPLATFORM

RUN apk update && apk add git curl bind-tools

COPY _build/bin/$TARGETPLATFORM/entry-point /bin/ep
COPY _build/bin/$TARGETPLATFORM/image-builder /image-builder

WORKDIR /
CMD ["/bin/ep", "/image-builder"]
