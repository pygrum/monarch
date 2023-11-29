#!/bin/bash

set -e

MONARCH_NET=monarch-net
MONARCH_PATH=${HOME}/.monarch

ACTIVE_CONTAINERS=$(docker network inspect \
  -f '{{ range $key, $value := .Containers }}{{ printf "%s\n" $key}}{{ end }}' ${MONARCH_NET})

if [ "$ACTIVE_CONTAINERS" ]; then
  echo "stopping and removing active containers on existing network"
  docker container stop $ACTIVE_CONTAINERS
  docker container rm $ACTIVE_CONTAINERS
fi
echo "removing monarch-net"
docker network rm "${MONARCH_NET}"

echo "purging configuration..."
rm -rf "$MONARCH_PATH"

echo "removing binaries"
rm -rf "${HOME}/.local/bin/monarch"
rm -rf "${HOME}/.local/bin/royal-lint"

echo "done"
