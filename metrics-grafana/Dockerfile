FROM --platform=$TARGETPLATFORM grafana/grafana:10.0.3
ARG TARGETPLATFORM

RUN grafana-cli plugins install grafana-piechart-panel

COPY metrics-grafana/grafana.ini /etc/grafana
COPY metrics-grafana/provisioning /etc/grafana/provisioning
COPY metrics-grafana/dashboards /etc/grafana/dashboards
COPY metrics-grafana/grafana.dark.3486e83352eeac079b82.css /usr/share/grafana/public/build/
