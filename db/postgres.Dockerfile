FROM golang:1.26.11-alpine AS build
RUN CGO_ENABLED=0 GOBIN=/out go install github.com/tianon/gosu@1.20

FROM postgres:16-alpine
RUN apk upgrade --no-cache
COPY --from=build /out/gosu /usr/local/bin/gosu
