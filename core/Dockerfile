FROM --platform=$TARGETPLATFORM alpine:3.17
ARG TARGETPLATFORM

RUN apk add curl tzdata lz4 procps git openssh-client
COPY _build/bin/$TARGETPLATFORM/entry-point /bin/ep
COPY _build/bin/$TARGETPLATFORM/core /bin/core
COPY _build/webui /public
COPY _build/modules /modules
COPY core/term/node_modules /public/term/node_modules

WORKDIR /data
CMD ["/bin/ep", "/bin/core"]
