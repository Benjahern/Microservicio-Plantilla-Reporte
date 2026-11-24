FROM golang:1.26.11-alpine AS build
RUN CGO_ENABLED=0 GOBIN=/out go install -tags 'no_mysql no_sqlite3 no_ydb no_mssql no_clickhouse no_vertica no_libsql' github.com/pressly/goose/v3/cmd/goose@v3.30.1

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/goose /goose
COPY db/migrations /migrations
ENTRYPOINT ["/goose", "-dir", "/migrations"]
CMD ["up"]
