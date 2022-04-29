ARG GO_VERSION=1.18

# Build the app
FROM quay.io/projectquay/golang:${GO_VERSION} AS build
WORKDIR /build/
ADD . /build/
RUN go build -o local-clair ./cmd/cli

# Final image
FROM registry.access.redhat.com/ubi8/ubi-minimal as final
RUN microdnf install --disablerepo=* --enablerepo=ubi-8-baseos --enablerepo=ubi-8-appstream sqlite
ARG REBUILD_DB
COPY --from=build /build/local-clair /bin/local-clair
RUN if [ "$REBUILD_DB" = "1" ] ; then DB_PATH=/matcher /bin/local-clair update ; fi

