#!/bin/bash

# Variables
DOCKER_USER="mcolussi"
IMAGE_NAME_GOLC="golc"
IMAGE_NAME_RESULTSALL="resultsall"
VERSION="1.0.3"

TAG="amd64"
TAG1="arm64"

cd Docker

# Build images amd64
podman build -t ${IMAGE_NAME_GOLC}:${TAG} -f Dockerfile.golc.amd64 .
podman build -t ${IMAGE_NAME_RESULTSALL}:${TAG} -f Dockerfile.ResultsAll.amd64 .

# Build images arm64
podman build -t ${IMAGE_NAME_GOLC}:${TAG1} -f Dockerfile.golc.arm64 .
podman build -t ${IMAGE_NAME_RESULTSALL}:${TAG1} -f Dockerfile.ResultsAll.arm64 .

# Login to Docker Hub
podman login docker.io

# Tag images amd
podman tag ${IMAGE_NAME_GOLC}:${TAG} ${DOCKER_USER}/${IMAGE_NAME_GOLC}:${TAG}-${VERSION}
podman tag ${IMAGE_NAME_RESULTSALL}:${TAG} ${DOCKER_USER}/${IMAGE_NAME_RESULTSALL}:${TAG}-${VERSION}

# Tag images arm
podman tag ${IMAGE_NAME_GOLC}:${TAG1} ${DOCKER_USER}/${IMAGE_NAME_GOLC}:${TAG1}-${VERSION}
podman tag ${IMAGE_NAME_RESULTSALL}:${TAG1} ${DOCKER_USER}/${IMAGE_NAME_RESULTSALL}:${TAG1}-${VERSION}

# Push images amd
podman push ${DOCKER_USER}/${IMAGE_NAME_GOLC}:${TAG}-${VERSION}
podman push ${DOCKER_USER}/${IMAGE_NAME_RESULTSALL}:${TAG}-${VERSION}

# Push images arm
podman push ${DOCKER_USER}/${IMAGE_NAME_GOLC}:${TAG1}-${VERSION}
podman push ${DOCKER_USER}/${IMAGE_NAME_RESULTSALL}:${TAG1}-${VERSION}