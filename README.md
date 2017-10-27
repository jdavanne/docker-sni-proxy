# docker-sni-proxy

Simplistic reverse proxy to route HTTPS(443) and HTTP(80) connection to actual dockers

# How to use

## CLI / Environment options
```sh
Usage of ./docker-sni-proxy:
  -docker-network string
    	Specify the public network for docker
  -host string
    	Specify the interface to listen to. (default "0.0.0.0")
  -http-port int
    	Specify the port to listen to. (default 80)
  -mode string
    	Specify the mode : stack, service (default "stack")
  -tls-port int
    	Specify the port to listen to. (default 443)
```

## Use cases
### With stack network within a unique compose/swarm stack, manual network assignment
- Use `MODE=service` or `MODE=stack` mode
- Add public service within `docker-sni-proxy` network
- see [./docker-compose.test.yml]

### With external public network, with multiple compose/swarm stacks, manual network assignment
- Use preferably `MODE=stack` mode
- Add public service to external public network

### With external public network, with multiple compose/swarm stacks, automatic network assignment
- Use preferably `MODE=stack` mode
- Use `DOCKER_NETWORK=<public-network-name>`
- see [docker-compose.yml]

## Modes
### Service mode
- Integrate `docker-sni-proxy` within your services with `MODE=service` (see [docker-compose.test.yml])
- To call a service <SERVICE>, use <SERVICE>.<DOMAIN>, `docker-sni-proxy` will call <SERVICE>

### Stack mode
- Integrate `docker-sni-proxy` in a dedicated stack with `MODE=stack`
- manual : Ensure all other public service are linked to that public network `<public-network-name>`
- automatic : use `DOCKER_NETWORK=<public-network-name>` (see [docker-compose.yml])
- To call a service <SERVICE> within the stack <STACK>, use <SERVICE>.<STACK>.<DOMAIN>, `docker-sni-proxy` will call `STACK_SERVICE`

## Changelog
- 0.3.0
  - add support for docker swarm mode service
- 0.2.0
  - add docker support : `docker-network=public` for the proxy, `labels: proxy=true` for containers/services
  - add docker internal port mapping support for http (80): `labels: proxy-http-port=<port>`
  - add docker internal port mapping support for tls/sni (443): `labels: proxy-tls-port=<port>`
- 0.0.3
  - Fix HTTP header Host parsing
- 0.0.2
  - Add support to HTTP on port 80
  - Add mode `service`. default remains `stack`
- 0.0.1
  - Basic support of SNI on port 443
