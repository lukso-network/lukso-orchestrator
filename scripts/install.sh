#!/bin/sh
RELEASE_TAG=v0.0.28-alpha;
sudo wget https://github.com/lukso-network/lukso-orchestrator/releases/download/$RELEASE_TAG/celebrimbor -O /usr/local/bin/celebrimbor
sudo chmod +x /usr/local/bin/celebrimbor
echo ;
