#!/bin/bash
docker login -u $DOCKERUSER -p $DOCKERPASS

DOCKER_CONTENT_TRUST=1 docker build . -t bsycorp/km:${TRAVIS_BRANCH//\//-}
DOCKER_CONTENT_TRUST=1 docker build . -f Dockerfile-slim -t bsycorp/km:${TRAVIS_BRANCH//\//-}-slim

docker push bsycorp/km