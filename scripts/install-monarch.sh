#!/bin/bash

set -e

cd "$(dirname "$0")"

MONARCH_NAME="monarch"
MONARCH_NET=monarch-net
MONARCH_PATH=${HOME}/.monarch

if [ -d "${MONARCH_PATH}" ]
then
  read -p "monarch folder exists. do you wish to reinstall? (y/N) " yn
  if [ "$yn" != "y" ] && [ "$yn" != "Y" ]; then
    exit 0
  fi
  rm -rf "${MONARCH_PATH}" && mkdir "${MONARCH_PATH}"
  ACTIVE_CONTAINERS=$(docker network inspect \
    -f '{{ range $key, $value := .Containers }}{{ printf "%s\n" $key}}{{ end }}' \
    ${MONARCH_NET})

  if [ "$ACTIVE_CONTAINERS" ]; then
    echo "stopping and removing active containers on existing network"
    docker container stop "$ACTIVE_CONTAINERS"
    docker container rm "$ACTIVE_CONTAINERS"
  fi
  docker network rm "${MONARCH_NET}"
else
  mkdir "${MONARCH_PATH}"
fi

if ! command -v docker &> /dev/null
then
  echo "please install docker and start the daemon."
  exit 1
fi

echo "creating docker network ${MONARCH_NET}"
docker network create "${MONARCH_NET}" --subnet 172.20.0.0/16

docker pull mysql:latest

echo "starting sql container"
docker run -dit --network ${MONARCH_NET} --ip 172.20.0.3 -e "MYSQL_ROOT_PASSWORD=monarch" \
  -e "MYSQL_DATABASE=monarch" --restart unless-stopped --name monarch-sql mysql:latest

cp ../configs/monarch.yaml "${MONARCH_PATH}"

echo "generating self-signed certs ${MONARCH_NAME}.crt ${MONARCH_NAME}.key..."
openssl req -newkey rsa:4096 \
            -x509 \
            -sha256 \
            -days 3650 \
            -nodes \
            -out "${MONARCH_PATH}/${MONARCH_NAME}.crt" \
            -keyout "${MONARCH_PATH}/${MONARCH_NAME}.key" \
            -subj "/C=US/ST=New York/L=New York City/O=${MONARCH_NAME}/OU=${MONARCH_NAME}/CN=www.${MONARCH_NAME}.com"

echo "done"

echo "building monarch.."
mkdir -p "${HOME}/.local/bin" 2>/dev/null

go build ../cmd/monarch/monarch.go
mv ./monarch "${HOME}/.local/bin"

echo "done. please wait for all services to start"