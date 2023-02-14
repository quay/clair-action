FROM quay.io/projectquay/clair-action:v0.0.4

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
