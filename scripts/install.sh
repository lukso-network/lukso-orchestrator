#!/bin/sh
RELEASE_TAG=v0.0.28-beta;
sudo wget https://github.com/lukso-network/lukso-orchestrator/releases/download/$RELEASE_TAG/lukso -O /usr/local/bin/lukso &&
sudo wget https://raw.githubusercontent.com/lukso-network/lukso-orchestrator/feature/47/installation-script/scripts/lukso-cli -O /usr/local/bin/lukso-cli &&
sudo chmod +x /usr/local/bin/lukso &&
sudo chmod +x /usr/local/bin/lukso-cli &&
echo "Type lukso-cli to start the node";
