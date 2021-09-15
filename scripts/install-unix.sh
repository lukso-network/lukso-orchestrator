#!/usr/bin/env bash

NETWORK="l15"
REPOSITORY="https://storage.googleapis.com/l16-common/l15-cdn";
PLATFORM="unknown";

ORCHESTRATOR_TAG="";
PANDORA_TAG="";
VANGUARD_TAG="";

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

download() {
  URL="$1";
  LOCATION="$2";
  if [ $PLATFORM == "Linux" ]; then
    sudo wget -O $LOCATION $URL;
  fi

  if [ $PLATFORM == "Darwin" ]; then
    sudo curl --output $LOCATION $URL;
  fi
}



sudo mkdir \
/opt/lukso \
/opt/lukso/tmp \
/opt/lukso/binaries \
/opt/lukso/networks \
/opt/lukso/networks/"$NETWORK" \
/opt/lukso/networks/"$NETWORK"/config;

download /home/pk/projects/lukso/lukso-orchestrator/scripts/install-unix.sh /opt/lukso/lukso;
download "$REPOSITORY"/config.zip /opt/lukso/tmp/config.zip;

sudo unzip /opt/lukso/tmp/config.zip -d /opt/lukso/networks/"$NETWORK"/config;

sudo chmod +x /opt/lukso/lukso;
sudo ln -sfn /opt/lukso/lukso /usr/local/bin/lukso;

sudo rm -rf /opt/lukso/tmp;

echo "Ready! type lukso to start the node!";
