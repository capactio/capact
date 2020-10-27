#!/bin/bash
echo "SKIP_DEPS_INSTALLATION=false" >>$GITHUB_ENV
echo "GO_VERSION=^1.15.2" >>$GITHUB_ENV
echo "PROJECT_ID=projectvoltron" >>$GITHUB_ENV
echo "LINT_TIMEOUT=2m" >>$GITHUB_ENV


if [ "${GITHUB_EVENT_NAME}" = "push" ]
then
  echo "DOCKER_TAG=$(echo ${GITHUB_SHA:0:7})" >>$GITHUB_ENV
else
  PR_NUMBER=$(echo $GITHUB_REF | awk 'BEGIN { FS = "/" } ; { print $3 }')
  echo "DOCKER_TAG=PR-$(echo ${PR_NUMBER})" >>$GITHUB_ENV
  echo "DOCKER_PUSH_REPOSITORY=gcr.io/projectvoltron/pr" >>$GITHUB_ENV
fi

echo APPS='name=matrix::{"include":[{"APP":"gateway"},{"APP":"k8s-engine"},{"APP":"och"}]}' >>$GITHUB_ENV
echo TESTS='name=matrix::{"include":[{"TEST":"e2e"}]}' >>$GITHUB_ENV
echo INFRAS='name=matrix::{"include":[{"INFRA":"json-go-gen"},{"INFRA":"graphql-schema-linter"}]}' >>$GITHUB_ENV


