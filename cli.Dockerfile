ARG GO_VERSION=1.23

# Build the app
FROM quay.io/projectquay/golang:${GO_VERSION} AS build
ARG CLAIR_ACTION_VERSION=""
WORKDIR /build/
ADD . /build/
RUN go build \
    -o clair-action \
    -trimpath \
    -ldflags="-s -w$(test -n "${CLAIR_ACTION_VERSION}" && printf " -X 'main.Version=%s'" "${CLAIR_ACTION_VERSION}")" \
    ./cmd/clair-action

# Final image
FROM registry.access.redhat.com/ubi8/ubi-minimal as final
COPY --from=build /build/clair-action /bin/clair-action
