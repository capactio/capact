#!/usr/bin/env bash

echo "Setting up CI environmental variables..."

cat <<EOT >> "$GITHUB_ENV"
GO_VERSION=^1.15.2
SKIP_DEPS_INSTALLATION=false
PROJECT_ID=projectvoltron
EOT

if [ "${GITHUB_EVENT_NAME}" = "push" ]
then
  echo "DOCKER_TAG=${GITHUB_SHA:0:7}" >> "$GITHUB_ENV"
else
  PR_NUMBER=$(echo "$GITHUB_REF" | awk 'BEGIN { FS = "/" } ; { print $3 }')
  echo "DOCKER_TAG=PR-${PR_NUMBER}" >> "$GITHUB_ENV"
  echo "DOCKER_REPOSITORY=gcr.io/projectvoltron/pr" >> "$GITHUB_ENV"
fi

# TODO: Read components to build in automated way, e.g. from directory structure
cat <<EOT >> "$GITHUB_ENV"
APPS=name=matrix::{"include":[{"APP":"gateway"},{"APP":"k8s-engine"},{"APP":"och"}]}
TESTS=name=matrix::{"include":[{"TEST":"e2e"}]}
INFRAS=name=matrix::{"include":[{"INFRA":"json-go-gen"},{"INFRA":"graphql-schema-linter"}]}
EOT

