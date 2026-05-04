#!/usr/bin/env bash
set -euo pipefail

# Rebuilds the current two-location server and restarts the systemd service.
git pull --ff-only
go build -o server ./cmd/server
sudo systemctl restart weather-server
echo "Deployed successfully."
