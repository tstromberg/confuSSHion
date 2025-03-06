#!/bin/sh
PROJECT="canary-452112"
export KO_DOCKER_REPO="gcr.io/${PROJECT}/canary"

gcloud run deploy pipeline-notifier --image="$(ko publish .)" --args=-serve \
  --region us-central1 --project "${PROJECT}"
