FROM node:24-alpine AS page-build
WORKDIR /src
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/web/tsconfig.app.json apps/web/tsconfig.app.json
COPY apps/web/status/package.json apps/web/status/package.json
COPY packages/status-contracts-web/package.json packages/status-contracts-web/package.json
COPY packages/web-ui/package.json packages/web-ui/package.json
RUN pnpm install --frozen-lockfile --filter @realtime-me/status-web...
COPY apps/web/status apps/web/status
COPY packages/status-contracts-web packages/status-contracts-web
COPY packages/web-ui packages/web-ui
RUN pnpm --filter @realtime-me/status-web build

FROM golang:1.26-alpine AS build
ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor vendor
COPY gen/go gen/go
COPY services/status services/status
RUN CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags="-s -w" -o /out/status-service ./services/status

FROM alpine:3.23
RUN adduser -D -H -u 10001 appuser \
    && mkdir -p /data \
    && chown appuser:appuser /data
COPY --from=build /out/status-service /usr/local/bin/status-service
COPY --from=page-build /src/apps/web/status/dist /app/status-web
ENV STATUS_WEB_DIR=/app/status-web
USER appuser
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/status-service"]
