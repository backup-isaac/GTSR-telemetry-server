FROM nginx

COPY domain-resolver.sh /usr/local/bin/domain-resolver.sh
COPY allowed-domain.list /etc/nginx/allowed-domain.list
COPY update-dns.sh /usr/local/bin/update-dns.sh
RUN apt-get update && apt-get install -y cron && chmod +x /usr/local/bin/domain-resolver.sh && chmod +x /usr/local/bin/update-dns.sh
RUN echo "0 1 * * * root /usr/local/bin/update-dns.sh" > /etc/cron.d/update-dns && domain-resolver.sh /etc/nginx/allowed-domain.list > /etc/nginx/allowed-ips-from-domains.conf && update-rc.d cron enable
