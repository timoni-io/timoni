FROM --platform=$TARGETPLATFORM traefik:v2.9.6
ARG TARGETPLATFORM

COPY _build/bin/$TARGETPLATFORM/entry-point /bin/ep

WORKDIR /
CMD ["/bin/ep", \
    "traefik", \
    "--metrics", \
    "--metrics.prometheus=true", \
    "--metrics.prometheus.addServicesLabels=true", \
    "--accesslog", \
    "--accesslog.format=json", \
    "--entryPoints.websecure.http.tls", \
    "--entryPoints.web.transport.respondingTimeouts.writeTimeout=30", \
    "--log", \
    "--log.level=INFO", \
    "--log.format=json", \
    "--api.dashboard", \
    "--providers.kubernetescrd", \
    "--providers.kubernetesingress", \
    "--entrypoints.web.address=:80", \
    "--entrypoints.websecure.address=:443", \
    "--entrypoints.traefik.address=:9000" \
]
