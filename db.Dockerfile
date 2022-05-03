ARG GO_VERSION=1.18

# Build the app
FROM quay.io/projectquay/golang:${GO_VERSION} AS build
WORKDIR /build/
ADD . /build/
RUN go build -o clair-action ./cmd/cli

# Final image
FROM registry.access.redhat.com/ubi8/ubi-minimal as final
RUN microdnf install --disablerepo=* --enablerepo=ubi-8-baseos --enablerepo=ubi-8-appstream sqlite unzip zstd
ARG REBUILD_DB
COPY --from=build /build/clair-action /bin/clair-action
RUN DB_PATH=/matcher /bin/clair-action update
RUN echo $(md5sum /matcher | awk '{ print $1 }') > /matcher_checksum.txt 
RUN zstd /matcher

# Do S3 upload
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
    unzip -q awscliv2.zip && \
    ./aws/install

RUN --mount=type=secret,id=aws \
    AWS_SHARED_CREDENTIALS_FILE=/run/secrets/aws \
    aws s3 cp --quiet --region=us-east-1 \
    /matcher.zst s3://clair-sqlite-db/matcher.zst \
    --metadata='{"checksum":"'$(cat /matcher_checksum.txt)'","date":"'$(date +%Y%m%d)'","version":"v0"}'
