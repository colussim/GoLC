#!/bin/bash


# Variables
DOCKER_USER="XXXXXXXXXX"
IMAGE_NAME_GOLC="golc"
IMAGE_NAME_RESULTSALL="resultsall"
VERSION="1.0.6"
USERNAME="XXXXXXXXXXXXXXXXXXXXXXXX"
REPOS="GoLC"
CMD="podman"

DIR="Docker/v${VERSION}"

TAG="amd64"
TAG1="arm64"

if [ -d "$DIR" ]; then
  echo "Directory $DIR exists. Deleting it."
  rm -rf "$DIR"
else
  echo "Directory $DIR does not exist. Creating it."
fi

mkdir -p ${DIR}
mkdir -p ${DIR}/golc
mkdir -p ${DIR}/resultsall
cp  Docker/*.amd64 ${DIR}/
cp  Docker/*.arm64 ${DIR}/

cp Release/v${VERSION}/arm64/linux/golc_v${VERSION}_linux_arm64/golc ${DIR}/golc/golc_arm64
cp Release/v${VERSION}/arm64/linux/golc_v${VERSION}_linux_arm64/ResultsAll ${DIR}/resultsall/ResultsAll_arm64
cp -r Release/v${VERSION}/arm64/linux/golc_v${VERSION}_linux_arm64/dist ${DIR}/resultsall/
cp -r Release/v${VERSION}/arm64/linux/golc_v${VERSION}_linux_arm64/imgs ${DIR}/resultsall/

cp Release/v${VERSION}/amd64/linux/golc_v${VERSION}_linux_amd64/golc ${DIR}/golc/golc_amd64
cp Release/v${VERSION}/amd64/linux/golc_v${VERSION}_linux_amd64/ResultsAll ${DIR}/resultsall/ResultsAll_amd64
cp -r Release/v${VERSION}/amd64/linux/golc_v${VERSION}_linux_amd64/dist ${DIR}/resultsall/
cp -r Release/v${VERSION}/amd64/linux/golc_v${VERSION}_linux_amd64/imgs ${DIR}/resultsall/

cd ${DIR}

# Build images amd64
${CMD} build -t ${IMAGE_NAME_GOLC}:${TAG} -f Dockerfile.golc.amd64 .
${CMD} build -t ${IMAGE_NAME_RESULTSALL}:${TAG} -f Dockerfile.ResultsAll.amd64 .


# Build images arm64
${CMD} build -t ${IMAGE_NAME_GOLC}:${TAG1} -f Dockerfile.golc.arm64 .
${CMD} build -t ${IMAGE_NAME_RESULTSALL}:${TAG1} -f Dockerfile.ResultsAll.arm64 .


# Login to Docker Hub
${CMD} login docker.io

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



