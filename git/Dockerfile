FROM scratch
ARG TARGETPLATFORM

# git server
COPY _build/bin/$TARGETPLATFORM/git /git

WORKDIR /
CMD ["/git"]
