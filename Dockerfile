ARG GO_VERSION=1.18

# Build the app
FROM quay.io/projectquay/golang:${GO_VERSION} AS build
WORKDIR /build/
ADD . /build/
RUN go build -o local-clair ./cmd/cli

# Pull latest vuln DB 
FROM quay.io/crozzy/clair-sqlite-db:latest as db
RUN gzip -1 /matcher

# Final image
FROM registry.access.redhat.com/ubi8/ubi-minimal AS final
RUN microdnf install --disablerepo=* --enablerepo=ubi-8-baseos --enablerepo=ubi-8-appstream gzip sqlite
ARG REBUILD_DB
COPY --from=build /build/local-clair /bin/local-clair
COPY --from=db /matcher.gz /matcher.gz
RUN gzip -d /matcher.gz
RUN if [ "$REBUILD_DB" = "1" ] ; then DB_PATH=/matcher /bin/local-clair update ; fi

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
