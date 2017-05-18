#!/bin/bash
#
set -euo pipefail

test_up() {
  docker-compose -f docker-compose.yml build
  docker-compose -f docker-compose.app.yml build

  docker stack deploy -c docker-compose.yml public
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
  echo "Test $1 OK ($2 = $3)"
}

test_run() {
  expect "app1.stack1" "$(curl --fail -s -k https://app1.stack1.localtest.me:443/)" "=stack1_app1="
  expect "app2.stack1" "$(curl --fail -s -k https://app2.stack1.localtest.me:443/)" "=stack1_app2="
  expect "app1.stack2" "$(curl --fail -s -k https://app1.stack2.localtest.me:443/)" "=stack2_app1="
  expect "app2.stack2" "$(curl --fail -s -k https://app2.stack2.localtest.me:443/)" "=stack2_app2="
  echo "All tests passed!"
}

case ${1:-} in
  run)
    test_clean
    test_up
    sleep 4
    test_run
    test_clean
  ;;

  test-only)
    test_run
  ;;

  up)
    test_clean
    test_up
  ;;

  clean)
    test_clean
  ;;

  *)
    echo "$0 usage: run | test-only | up | clean"
  ;;
esac
