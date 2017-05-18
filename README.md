# docker-sni-proxy
#

Simplistic reverse proxy to route HTTPS(443) connection to actual dockers

# How to use

- launch the sni reverse proxy as a stack "public" creating a "default" network `public_default`:
```bash
docker stack deploy -c docker-compose.yml public
```

- launch an applicative stack "stack1"
```bash
STACK=stack1 docker stack deploy -c docker-compose.app.yml stack1
```

- launch another applicative stack "stack2"
```bash
STACK=stack2 docker stack deploy -c docker-compose.app.yml stack1
```

- now test whether it works
```
$ curl -k https://app1.stack1.localtest.me:443/
=stack1_app1=
$ curl -k https://app2.stack1.localtest.me:443/
=stack1_app2=


$ curl -k https://app1.stack2.localtest.me:443/
=stack2_app1=
$ curl -k https://app2.stack2.localtest.me:443/
=stack2_app2=

$ curl -k https://app3.stack1.localtest.me:443/
curl: (35) Unknown SSL protocol error in connection to app3.stack1.localtest.me:443
```
