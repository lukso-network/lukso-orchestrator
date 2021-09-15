#!/usr/bin/env bash

NETWORK="l15"
REPOSITORY="https://storage.googleapis.com/l16-common/l15-cdn";
PLATFORM="unknown";
ARCHITECTURE=$(uname -m);

ORCHESTRATOR_TAG="";
PANDORA_TAG="";
VANGUARD_TAG="";

if [ "$OSTYPE" = "linux-gnu" ]; then
  PLATFORM="Linux";
elif [[ "$OSTYPE" = "darwin"* ]]; then
  PLATFORM="Darwin"
elif [ "$OSTYPE" = "cygwin" ]; then
  PLATFORM="Linux"
elif [ "$OSTYPE" = "freebsd" ]; then
  PLATFORM="Linux"
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

download https://raw.githubusercontent.com/lukso-network/lukso-orchestrator/feature/l15-setup-script/scripts/lukso /opt/lukso/lukso;

download "$REPOSITORY"/config.zip /opt/lukso/tmp/config.zip;

sudo unzip /opt/lukso/tmp/config.zip -d /opt/lukso/networks/"$NETWORK"/config;

sudo chmod +x /opt/lukso/lukso;
sudo ln -sfn /opt/lukso/lukso /usr/local/bin/lukso;

sudo rm -rf /opt/lukso/tmp;

sudo lukso bind-binaries \
--orchestrator v0.1.0-beta.1 \
--pandora v0.1.0-beta.1 \
--vanguard v0.1.0-beta.1 \
--validator v0.1.0-beta.1;

echo "Ready! type lukso to start the node!";
