FROM quay.io/crozzy/clair-action:v0.0.1

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
