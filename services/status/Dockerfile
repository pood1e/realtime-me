FROM golang:1.26-alpine AS build
ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor vendor
COPY gen/go gen/go
COPY libs libs
COPY services/status services/status
RUN CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags="-s -w" -o /out/status-service ./services/status

FROM alpine:3.23
RUN adduser -D -H -u 10001 appuser \
    && mkdir -p /data \
    && chown appuser:appuser /data
COPY --from=build /out/status-service /usr/local/bin/status-service
USER appuser
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/status-service"]
