#!/usr/bin/env bash

NETWORK="l15"
REPOSITORY="https://storage.googleapis.com/l16-common/l15-cdn";
PLATFORM="unknown";

if [ "$OSTYPE" = "linux-gnu" ]; then
  PLATFORM="linux";
elif [ "$OSTYPE" = "darwin" ]; then
  PLATFORM="darwin"
elif [ "$OSTYPE" = "cygwin" ]; then
  PLATFORM="linux"
elif [ "$OSTYPE" = "msys" ]; then
  PLATFORM="windows"
elif [ "$OSTYPE" = "win32" ]; then
  PLATFORM="windows"
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
/opt/lukso/networks \
/opt/lukso/networks/"$NETWORK" \
/opt/lukso/networks/"$NETWORK"/bin \
/opt/lukso/networks/"$NETWORK"/config;


if [ "$PLATFORM" = "linux" ]; then
  sudo wget -O /opt/lukso/tmp/bin.zip "$REPOSITORY"/linux-binaries.zip;
fi

if [ "$PLATFORM" = "darwin" ]; then
  sudo wget -O /opt/lukso/tmp/bin.zip "$REPOSITORY"/darwin-binaries.zip;
fi

sudo wget -O /opt/lukso/tmp/config.zip "$REPOSITORY"/config.zip;

sudo unzip /opt/lukso/tmp/config.zip -d /opt/lukso/networks/"$NETWORK"/config;
sudo unzip /opt/lukso/tmp/bin.zip -d /opt/lukso/networks/"$NETWORK"/bin;

sudo chmod -R +x /opt/lukso/networks/"$NETWORK"/bin;

sudo ln -s /opt/lukso/networks/"$NETWORK"/bin/lukso /usr/local/bin/lukso;

sudo rm -rf /opt/lukso/tmp;

echo "Ready! type lukso to start the node!";
