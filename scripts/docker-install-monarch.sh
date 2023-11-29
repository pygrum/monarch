#!/bin/bash

# This file is used to setup monarch within a docker container.
set -e

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