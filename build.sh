#!/bin/bash
set -e

echo "Building Cloud DR Orchestrator Docker image..."

# Build Docker image
docker build -t ghcr.io/kobeep/cloud-dr-orchestrator:latest .

echo "✅ Image built successfully!"
echo ""
echo "To push to GitHub Container Registry:"
echo "1. docker login ghcr.io"
echo "2. docker push ghcr.io/kobeep/cloud-dr-orchestrator:latest"
echo ""
echo "Or build and load locally into K3S:"
echo "docker save ghcr.io/kobeep/cloud-dr-orchestrator:latest | sudo k3s ctr images import -"
