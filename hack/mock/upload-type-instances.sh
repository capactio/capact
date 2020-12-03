#!/bin/sh

echo
echo - Creating Jira TypeInstance...
case "$1" in
helm)
  echo - Uploading jira-instance to OCH [ID: 45678917-ab4a-4fbf-a312-f96f11b07d0b]
  ;;
cloudsql)
  echo - Uploading jira-instance to OCH [ID: 13343627-ab4a-4fbf-a312-f96f11b07d0b]
  ;;
esac

sleep 5
