ARG GO_VERSION=1.18
ARG BUILD_DB
FROM quay.io/projectquay/golang:${GO_VERSION} AS build
RUN dnf install -y --disablerepo=* --enablerepo=ubi-8-baseos --enablerepo=ubi-8-appstream --enablerepo=ubi-8-codeready-builder gpgme-devel device-mapper-devel
WORKDIR /build/
ADD . /build/
RUN go build -tags exclude_graphdriver_btrfs,containers_image_openpgp .

FROM registry.access.redhat.com/ubi8/ubi-minimal AS final

# Needed for matching
RUN microdnf install --disablerepo=* --enablerepo=ubi-8-baseos --enablerepo=ubi-8-appstream sqlite device-mapper

COPY --from=build /build/local-clair /bin/local-clair
RUN DB_PATH=/matcher /bin/local-clair update
RUN if [ "$BUILD_DB" = "1" ] ; then DB_PATH=/matcher /bin/local-clair update ; fi
