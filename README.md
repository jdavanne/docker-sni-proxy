# docker-sni-proxy
#

Simplistic reverse proxy to route HTTPS(443) connection to actual dockers

# How to use

## Service mode
- Integrate docker-sni-proxy within your services with MODE=service
- To call a service SERVICE, use SERVICE.DOMAIN,
  docker-sni-proxy will call SERVICE


## Stack mode
- Integrate docker-sni-proxy in a dedicated stack with a "public" network.
- Ensure all other public service are link to that "public network"
- To call a service SERVICE within the stack STACK, use SERVICE.STACK.DOMAIN,
  docker-sni-proxy will call STACK_SERVICE

## Changelog
- 0.2.0
  - add docker support : `docker-network=public`, `labels: proxy=true`
  - add docker internal port mapping support for http (80): `labels: proxy-http-port=<port>`
  - add docker internal port mapping support for tls/sni (443): `labels: proxy-tls-port=<port>`
- 0.0.3
  - Fix HTTP header Host parsing
- 0.0.2
  - Add support to HTTP on port 80
  - Add mode `service`. default remains `stack`
- 0.0.1
  - Basic support of SNI on port 443
