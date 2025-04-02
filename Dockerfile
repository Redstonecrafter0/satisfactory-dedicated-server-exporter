FROM golang:1.24 AS build

WORKDIR /go/src/app
COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -ldflags "-w -s" -o /go/bin/app

FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.source="https://github.com/Redstonecrafter0/satisfactory-dedicated-server-exporter"
LABEL org.opencontainers.image.description="Satisfactory Dedicated Server Prometheus-Exporter"
LABEL org.opencontainers.image.licenses="AGPL-3.0-or-later"

COPY --from=build /go/bin/app /

ENV SDSE_PORT=7777

EXPOSE 8080

CMD ["/app"]
