#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ROOT_DIR=$(dirname "$SCRIPT_DIR")

CONTAINER_CLI="${CONTAINER_CLI:-podman}"
IMAGE_NAME="${IMAGE_NAME:-cardgame-server}"

# tag 規則：HEAD 有 git tag 用 git tag，否則用 latest
if [ -z "${IMAGE_TAG:-}" ]; then
    IMAGE_TAG=$(git -C "$ROOT_DIR" describe --tags --exact-match 2>/dev/null || echo latest)
fi

echo "Building image: $IMAGE_NAME:$IMAGE_TAG with $CONTAINER_CLI"
"$CONTAINER_CLI" build \
    -f "$SCRIPT_DIR/Dockerfile.server" \
    -t "$IMAGE_NAME:$IMAGE_TAG" \
    "$ROOT_DIR/Server"

echo "Build completed: $IMAGE_NAME:$IMAGE_TAG"
