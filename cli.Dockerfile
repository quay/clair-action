ARG GO_VERSION=1.18

# Build the app
FROM quay.io/projectquay/golang:${GO_VERSION} AS build
WORKDIR /build/
ADD . /build/
RUN go build -o clair-action ./cmd/cli

# Final image
FROM registry.access.redhat.com/ubi8/ubi-minimal as final
COPY --from=build /build/clair-action /bin/clair-action
