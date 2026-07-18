FROM golang:1.26.5-alpine AS build

WORKDIR /src/api
COPY api ./
RUN CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags='-s -w' -o /out/cloud-drive-api ./cmd/server \
	&& CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags='-s -w' -o /out/cloud-drive-worker ./cmd/worker \
	&& CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags='-s -w' -o /out/cloud-drive-migrate ./cmd/migrate

FROM alpine:3.22

RUN apk add --no-cache ca-certificates ffmpeg poppler-utils vips-tools

COPY --from=build /out/cloud-drive-api /usr/local/bin/cloud-drive-api
COPY --from=build /out/cloud-drive-worker /usr/local/bin/cloud-drive-worker
COPY --from=build /out/cloud-drive-migrate /usr/local/bin/cloud-drive-migrate

USER 10001:10001
EXPOSE 8080
CMD ["/usr/local/bin/cloud-drive-api"]
