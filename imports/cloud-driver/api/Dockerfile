FROM golang:1.26.5-alpine AS build

WORKDIR /src/api
COPY api ./
RUN CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags='-s -w' -o /out/cloud-drive-api ./cmd/server

FROM alpine:3.22

COPY --from=build /out/cloud-drive-api /usr/local/bin/cloud-drive-api

USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/cloud-drive-api"]
