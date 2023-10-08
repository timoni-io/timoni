FROM --platform=$TARGETPLATFORM alpine:3.17
ARG TARGETPLATFORM

RUN apk add curl procps
COPY _build/bin/$TARGETPLATFORM/entry-point /bin/ep
COPY _build/bin/$TARGETPLATFORM/node-agent /node-agent

WORKDIR /
CMD ["/bin/ep", "/node-agent"]
