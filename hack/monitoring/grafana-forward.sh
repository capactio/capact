#!/bin/bash

USERNAME=$(kubectl -n monitoring get secrets monitoring-grafana -ojsonpath="{.data.admin-user}" | base64 -d)
PASSWORD=$(kubectl -n monitoring get secrets monitoring-grafana -ojsonpath="{.data.admin-password}" | base64 -d)

echo "Username: $USERNAME"
echo "Password: $PASSWORD"

kubectl -n monitoring port-forward svc/monitoring-grafana 3000:80
