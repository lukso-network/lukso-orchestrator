#!/usr/bin/env bash

NETWORK="l15"
REPOSITORY="https://storage.googleapis.com/l16-common/l15-cdn";
PLATFORM="unknown";

if [ "$OSTYPE" = "linux-gnu" ]; then
  PLATFORM="linux";
elif [[ "$OSTYPE" = "darwin"* ]]; then
  PLATFORM="darwin"
elif [ "$OSTYPE" = "cygwin" ]; then
  PLATFORM="linux"
elif [ "$OSTYPE" = "freebsd" ]; then
  PLATFORM="linux"
fi

if [ "$PLATFORM" = "unknown" ]; then
  echo "Platform not supported.";
  exit;
fi

sudo mkdir \
/opt/lukso \
/opt/lukso/tmp \
/opt/lukso/binaries \
/opt/lukso/networks \
/opt/lukso/networks/"$NETWORK" \
/opt/lukso/networks/"$NETWORK"/config;


if [ "$PLATFORM" = "linux" ]; then
  sudo wget -O /opt/lukso/lukso https://raw.githubusercontent.com/lukso-network/lukso-orchestrator/feature/l15-setup/scripts/lukso;
  sudo wget -O /opt/lukso/tmp/config.zip "$REPOSITORY"/config.zip;
fi

if [ "$PLATFORM" = "darwin" ]; then
  sudo curl --output /opt/lukso/lukso https://raw.githubusercontent.com/lukso-network/lukso-orchestrator/feature/l15-setup/scripts/lukso;
  sudo curl --output /opt/lukso/tmp/config.zip "$REPOSITORY"/config.zip;
fi

sudo unzip /opt/lukso/tmp/config.zip -d /opt/lukso/networks/"$NETWORK"/config;

sudo chmod +x /opt/lukso/lukso;
sudo ln -sfn /opt/lukso/lukso /usr/local/bin/lukso;

sudo rm -rf /opt/lukso/tmp;

echo "Ready! type lukso to start the node!";
