# Caddy с плагином DuckDNS: сертификат получается через DNS, открывать порты 80 и 443 для проверки не нужно
FROM caddy:2-builder AS builder

RUN xcaddy build --with github.com/caddy-dns/duckdns

FROM caddy:2

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
