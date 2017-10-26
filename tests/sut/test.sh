#!/bin/sh
#
set -euo pipefail

DOMAIN=localtest.me

test_proxy() {
  docker-compose -f docker-compose.yml build
  docker stack deploy -c docker-compose.yml public
}

test_app() {
  docker-compose -f docker-compose.app.yml build

  sleep 1
  STACK=stack1 docker stack deploy -c docker-compose.app.yml stack1
  STACK=stack2 docker stack deploy -c docker-compose.app.yml stack2
}

test_clean() {
  docker stack rm stack1 || true
  docker stack rm stack2 || true
  docker stack rm public || true
}

expect() {
  [ ! "$2" = "$3" ] && (echo "Test $1 Failed $2 != $3" ; exit 1)
  echo "Test $1 OK ('$2' = '$3')"
}

getip() {
  case `uname -s` in
    Darwin)
      dscacheutil -q host -a name $1 | grep ip_address | awk '{ print $2 }'
    ;;
    *)
    getent hosts $1 | awk '{ print $1 }'
  esac
}

test_run() {
  ip=$(getip proxy)
  echo "ip: $ip"
  sleep 1
  curl -k --resolve app1.${DOMAIN}:443:${ip} https://app1.${DOMAIN}:443/
  expect "app1.stack1" "$(curl --fail -s -k --resolve app1.${DOMAIN}:443:${ip} https://app1.${DOMAIN}/)" "=app1="
  expect "app2.stack1" "$(curl --fail -s -k --resolve app2.${DOMAIN}:443:${ip} https://app2.${DOMAIN}/)" "=app2="
  expect "app1.stack2" "$(curl --fail -s -k --resolve app1.${DOMAIN}:443:${ip} https://app1.${DOMAIN}/)" "=app1="
  expect "app2.stack2" "$(curl --fail -s -k --resolve app2.${DOMAIN}:443:${ip} https://app2.${DOMAIN}/)" "=app2="
  expect "app4.stack1" "$(curl --fail -s -k --resolve app4.${DOMAIN}:80:${ip} http://app4.${DOMAIN}/)" "=app4="
  expect "app5.stack1" "$(curl --fail -s -k --resolve app5.${DOMAIN}:80:${ip} http://app5.${DOMAIN}/)" "=app5="
  expect "app6.stack1" "$(curl --fail -s -k --resolve app6.${DOMAIN}:443:${ip} https://app6.${DOMAIN}/)" "=app6="
  echo "All tests passed!"
}

case ${1:-} in
  run)
    test_clean
    test_proxy
    test_app
    sleep 4
    test_run
    test_clean
  ;;

  proxy)
    docker stack rm public || true
    test_proxy
  ;;

  test-only)
    test_run
  ;;

  up)
    test_clean
    test_proxy
    test_app
  ;;

  clean)
    test_clean
  ;;

  *)
    echo "$0 usage: run | test-only | up | clean"
  ;;
esac
