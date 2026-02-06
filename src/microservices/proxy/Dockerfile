FROM kong:3.6

USER root

RUN if command -v apk >/dev/null 2>&1; then \
      apk add --no-cache gettext; \
    else \
      apt-get update && apt-get install -y --no-install-recommends gettext-base && rm -rf /var/lib/apt/lists/*; \
    fi

COPY kong.yml.template /etc/kong/kong.yml.template
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

USER kong

ENV KONG_DATABASE=off
ENV KONG_DECLARATIVE_CONFIG=/etc/kong/kong.yml
ENV KONG_PROXY_LISTEN=0.0.0.0:8000
ENV KONG_ADMIN_LISTEN=0.0.0.0:8001

ENV KONG_NGINX_DAEMON=off


ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["kong", "start"]
