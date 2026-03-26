#!/bin/bash
set -e

if [ ! -f .env.deploy ]; then
  echo "✗ .env.deploy not found. Copy .env.deploy.example and fill in your values."
  exit 1
fi

echo "→ rebuilding binary..."
CGO_ENABLED=1 go build -o servicepatrol ./cmd/server/main.go

echo "→ rebuilding image..."
sudo podman build --no-cache -t docker.io/library/servicepatrol:latest .
sudo podman save docker.io/library/servicepatrol:latest | sudo k3s ctr images import -
rm servicepatrol

echo "→ updating config..."
sudo kubectl create configmap servicepatrol-config --from-env-file=.env.deploy -n servicepatrol --dry-run=client -o yaml | sudo kubectl apply -f -

sudo kubectl rollout restart deployment/servicepatrol -n servicepatrol
sudo kubectl rollout status deployment/servicepatrol -n servicepatrol

echo "✓ updated"
