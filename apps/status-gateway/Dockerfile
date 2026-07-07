FROM node:24-alpine AS page-build
WORKDIR /src
COPY package.json package-lock.json ./
COPY apps/status-page/package.json apps/status-page/package.json
RUN npm ci --workspace apps/status-page
COPY apps/status-page apps/status-page
WORKDIR /src/apps/status-page
RUN npm run build

FROM golang:1.26-alpine AS build
# GOPROXY is configurable so regional builds can use a reachable module mirror
# (e.g. --build-arg GOPROXY=https://goproxy.cn,direct).
ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}
WORKDIR /src
COPY apps/status-gateway/go.mod apps/status-gateway/go.mod
COPY apps/status-gateway apps/status-gateway
WORKDIR /src/apps/status-gateway
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/status-gateway .

FROM alpine:3.23
RUN adduser -D -H -u 10001 appuser
RUN mkdir -p /data && chown appuser:appuser /data
COPY --from=build /out/status-gateway /usr/local/bin/status-gateway
COPY --from=page-build /src/apps/status-page/dist /app/status-page
ENV STATUS_WEB_DIR=/app/status-page
USER appuser
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/status-gateway"]
