#!/bin/bash

set -e

cd "$(dirname "$0")"

MONARCH_NAME="monarch"
MONARCH_NET=monarch-net
MONARCH_PATH=${HOME}/.monarch
NET_EXISTS=$(docker network ls --filter name=^${MONARCH_NET}$ --format="{{ .Name }}")
if [ -d "${MONARCH_PATH}" ] || [ "$NET_EXISTS" ]
then
  read -p "monarch data exists. do you wish to reinstall? (y/N) " yn
  if [ "$yn" != "y" ] && [ "$yn" != "Y" ]; then
    exit 0
  fi
  if [ "$NET_EXISTS" ]
  then
    rm -rf "${MONARCH_PATH}" && mkdir "${MONARCH_PATH}"
    ACTIVE_CONTAINERS=$(docker network inspect \
      -f '{{ range $key, $value := .Containers }}{{ printf "%s\n" $key}}{{ end }}' \
      ${MONARCH_NET})

    if [ "$ACTIVE_CONTAINERS" ]; then
      echo "stopping and removing active containers on existing network"
      docker container stop $ACTIVE_CONTAINERS
      docker container rm $ACTIVE_CONTAINERS
    fi
  fi
  if [ "$(docker ps -qa -f name=monarch-ctr)" ]; then
    docker container rm monarch-ctr
  fi
  if [ "$NET_EXISTS" ]
  then
    docker network rm "${MONARCH_NET}"
  fi
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
cp ../configs/monarch_http.json "${MONARCH_PATH}"

echo "generating self-signed CA certificate ca-cert.pem ca-key.pem"

openssl req -x509 \
  -newkey rsa:4096 \
  -days 365 \
  -nodes \
  -keyout "${MONARCH_PATH}/ca-key.pem" \
  -out "${MONARCH_PATH}/ca-cert.pem" \
  -subj "/C=US/ST=New York/L=New York City/O=Organisation/OU=Education/CN=monarch"
echo "done"

echo "generating server key and certificate signing request"
openssl req -newkey rsa:4096 \
  -nodes \
  -out "${MONARCH_PATH}/server-req.pem" \
  -keyout "${MONARCH_PATH}/${MONARCH_NAME}-key.pem" \
  -subj "/C=US/ST=California/L=Mountain View/O=Google LLC/CN=*.google.com"

echo "signing server request with CA private key"
openssl x509 -req \
  -in "${MONARCH_PATH}/server-req.pem" \
  -days 60 \
  -CA "${MONARCH_PATH}/ca-cert.pem" \
  -CAkey "${MONARCH_PATH}/ca-key.pem" \
  -CAcreateserial \
  -out "${MONARCH_PATH}/${MONARCH_NAME}-cert.pem" \

rm -rf "${MONARCH_PATH}/server-req.pem"
echo "done"

mkdir -p "$HOME/.local/bin" 2>/dev/null

echo "installing linter for project configuration files"
go build ../cmd/linter/royal/royal-lint.go
mv royal-lint "$HOME/.local/bin/royal-lint"
echo "royal-lint saved to $HOME/.local/bin/royal-lint"

echo "building monarch.."
go build ../cmd/server/monarch-server.go
mv ./monarch-server "${HOME}/.local/bin"

echo "building monarch client..."
go build ../cmd/client/monarch-client.go
mv ./monarch-client "${HOME}/.local/bin"

echo "done. waiting for all services to start..."
sleep 5
echo "monarch-server saved to $HOME/.local/bin"
