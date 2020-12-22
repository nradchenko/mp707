#!/bin/bash

set -exo pipefail

mkdir -p "$(dirname "$OUTPUT")"

FROM=${FROM:-golang:1.15-buster}
GOARCH=${GOARCH:amd64}

DIGEST=$(docker manifest inspect "$FROM" | jq -r '.manifests[] | select(.platform.architecture == "'"$GOARCH"'") | .digest')
FROM=$FROM@$DIGEST

TAG=$SLUG:$GOARCH

docker build \
  --pull \
  -t "$TAG" \
  --build-arg FROM="$FROM" \
  .

CONTAINER=$(docker create "$TAG")

docker cp "$CONTAINER":/"$NAME" "$OUTPUT"
docker rm "$CONTAINER"
