FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
COPY vendor vendor
COPY services/library services/library
RUN CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags='-s -w' -o /out/library-api ./services/library/cmd/server \
	&& CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags='-s -w' -o /out/library-worker ./services/library/cmd/worker \
	&& CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags='-s -w' -o /out/library-migrate ./services/library/cmd/migrate

FROM alpine:3.23

RUN apk add --no-cache ca-certificates ffmpeg poppler-utils vips-tools

COPY --from=build /out/library-api /usr/local/bin/library-api
COPY --from=build /out/library-worker /usr/local/bin/library-worker
COPY --from=build /out/library-migrate /usr/local/bin/library-migrate

USER 10001:10001
EXPOSE 8080
CMD ["/usr/local/bin/library-api"]
