#!/bin/bash

# This file is used to setup monarch within a docker container.
set -e

MONARCH_NET=monarch-net
MONARCH_PATH=${HOME}/.monarch
NET_EXISTS=$(docker network ls --filter name=^${MONARCH_NET}$ --format="{{ .Name }}")
if [ -d "${MONARCH_PATH}" ] || [ "$NET_EXISTS" ]
then
  read -p "monarch data exists. do you wish to reinstall? (y/N) " yn
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

cd "$(dirname "$0")/.."

# build container
echo "building monarch container"
docker build -t monarch-ctr:latest -f docker/monarch/Dockerfile .
docker run -v /var/run/docker.sock:/var/run/docker.sock --name monarch-ctr -dit monarch-ctr:latest
echo "running installer on container..."
docker exec -i monarch-ctr bash -c 'chmod +x scripts/install-monarch.sh && ./scripts/install-monarch.sh'
docker exec -i monarch-ctr bash -c 'mv /root/.local/bin/monarch /usr/bin/monarch && echo "moved to /usr/bin/monarch"'
echo "done, connecting container to network"
docker network connect monarch-net monarch-ctr
echo "stopping container"
docker stop monarch-ctr
echo "done"

echo "creating run script..."
mkdir -p "$HOME/.local/bin" 2>/dev/null
cat <<EOF > "$HOME/.local/bin/monarch.sh"
#!/bin/bash
docker start monarch-ctr
docker exec -it monarch-ctr bash
EOF

chmod +x "$HOME/.local/bin/monarch.sh"
echo "done"
echo "monarch saved to $HOME/.local/bin/monarch.sh"
