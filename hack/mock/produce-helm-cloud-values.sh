#!/bin/sh

yq w -i /out/jira-cloud-values.yaml databaseConnection.host "$(yq r input.yaml host)"
yq w -i /out/jira-cloud-values.yaml databaseConnection.database "$(yq r input.yaml defaultDBName)"

sleep 5
