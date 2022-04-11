ARG GO_VERSION=1.18
FROM quay.io/projectquay/golang:${GO_VERSION} AS build
RUN dnf install -y --disablerepo=* --enablerepo=ubi-8-baseos --enablerepo=ubi-8-appstream --enablerepo=ubi-8-codeready-builder gpgme-devel device-mapper-devel
WORKDIR /build/
ADD . /build/
RUN go build -tags exclude_graphdriver_btrfs,containers_image_openpgp .

FROM registry.access.redhat.com/ubi8/ubi-minimal AS final

# Needed for matching
RUN microdnf install --disablerepo=* --enablerepo=ubi-8-baseos --enablerepo=ubi-8-appstream sqlite device-mapper

COPY --from=build /build/local-clair /bin/local-clair
COPY --from=build /build/matcher /matcher
