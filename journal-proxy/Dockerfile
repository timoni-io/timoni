FROM scratch
ARG TARGETPLATFORM

COPY _build/bin/$TARGETPLATFORM/journal-proxy /journal-proxy

WORKDIR /
CMD ["/journal-proxy"]
