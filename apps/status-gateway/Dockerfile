FROM golang:1.26-alpine AS build
WORKDIR /src
COPY apps/status-gateway/go.mod apps/status-gateway/go.mod
COPY apps/status-gateway apps/status-gateway
WORKDIR /src/apps/status-gateway
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/status-gateway .

FROM alpine:3.23
RUN adduser -D -H -u 10001 appuser
RUN mkdir -p /data && chown appuser:appuser /data
COPY --from=build /out/status-gateway /usr/local/bin/status-gateway
USER appuser
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/status-gateway"]
