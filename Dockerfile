FROM quay.io/crozzy/clair-sqlite-db:latest

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
