#!/usr/bin/env bash

# Couchbase docker run.

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
DOCKERFILE="${DIR}/../deploy/docker/Dockerfile.couchbase"
VERSION=${COUCHDB_TAG:-"latest"}
TAG_NAME="messageboxdb:${VERSION}"
CONTAINER_NAME="MessageBoxDB"
NETWORK=$1
echo "Building couchbase image"
docker build -f "${DOCKERFILE}" "${DIR}" -t "${TAG_NAME}"

if [[ $? -eq 0 ]]; then
  docker run --name ${CONTAINER_NAME} --net="${NETWORK}" -d -p 8091:8091 -p 8092:8092 -p 8093:8093 -p 11210:11210  "${TAG_NAME}"
fi
