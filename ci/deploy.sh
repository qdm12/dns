#!/bin/bash

function build {
  PLATFORM=$1
  BASE_IMAGE=$2
  TAG=$3
  echo "building ${IMAGE_PATH}:${TAG}..."
  buildctl build \
    --frontend dockerfile.v0 \
    --opt filename=./Dockerfile \
    --opt platform=linux/${PLATFORM} \
    --opt build-arg:BASE_IMAGE=$2 \
    --opt build-arg:BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
    --opt build-arg:VCS_REF=${TRAVIS_COMMIT} \
    --exporter image \
    --exporter-opt name=docker.io/${IMAGE_PATH}:${TAG} \
    --exporter-opt push=true \
    --local dockerfile=. \
    --local context=. &> /dev/null
  echo "built ${IMAGE_PATH}:${TAG} with exit code $?"
}

build amd64 alpine ${BASE_TAG}
build amd64 alpine ${BASE_TAG}-amd64
build 386 i386/alpine ${BASE_TAG}-i386
build arm/v6 arm32v6/alpine ${BASE_TAG}-arm32v6
build arm arm32v7/alpine ${BASE_TAG}-arm32v7
build arm64 arm64v8/alpine ${BASE_TAG}-arm64v8
docker manifest create ${IMAGE_PATH}:${BASE_TAG} \
    ${IMAGE_PATH}:${BASE_TAG}-amd64 \
    ${IMAGE_PATH}:${BASE_TAG}-i386 \
    ${IMAGE_PATH}:${BASE_TAG}-arm32v6 \
    ${IMAGE_PATH}:${BASE_TAG}-arm32v7 \
    ${IMAGE_PATH}:${BASE_TAG}-arm64v8
docker manifest annotate ${IMAGE_PATH}:${BASE_TAG} ${IMAGE_PATH}:${BASE_TAG}-amd64 --arch amd64
docker manifest annotate ${IMAGE_PATH}:${BASE_TAG} ${IMAGE_PATH}:${BASE_TAG}-arm32v6 --arch arm/v6
docker manifest annotate ${IMAGE_PATH}:${BASE_TAG} ${IMAGE_PATH}:${BASE_TAG}-arm32v7 --arch arm
docker manifest annotate ${IMAGE_PATH}:${BASE_TAG} ${IMAGE_PATH}:${BASE_TAG}-arm64v8 --arch arm64
docker manifest push ${IMAGE_PATH}:${BASE_TAG}
