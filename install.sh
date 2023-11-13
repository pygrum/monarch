#!/bin/bash

set -e

cd "$(dirname "$0")"

MONARCH_DOCKER_NET=monarch-net
MONARCH_PATH=${HOME}/.monarch

if [ -d "${MONARCH_PATH}" ]
then
  echo "monarch folder exists. skipping creation"
else
  mkdir "${MONARCH_PATH}"
fi

if ! command -v docker &> /dev/null
then
  echo "please install docker and start the daemon."
  exit 1
fi

echo "creating docker network ${MONARCH_DOCKER_NET} --subnet 172.20.0.0/16"
docker network create "${MONARCH_DOCKER_NET}"

docker pull mysql:latest

echo "starting sql container"
docker run --rm -ditp 3306:3306 --network ${MONARCH_DOCKER_NET} --ip 172.20.0.2 -e "MYSQL_ROOT_PASSWORD=monarch" \
  -e "MYSQL_DATABASE=monarch" mysql:latest

mv configs/.monarch.yaml "${MONARCH_PATH}"