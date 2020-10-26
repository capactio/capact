## Service Monitor 

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: "controller-manager-metrics-monitor"
  labels:
    prometheus: scrape
    {{- include "engine.labels" . | nindent 4 }}
spec:
  endpoints:
    - port: http-metrics
  namespaceSelector:
    matchNames:
      - "{{ .Release.Namespace }}"
  selector:
    matchLabels:
      {{- include "engine.labels" . | nindent 6 }}
```

## Apps

```bash
kubectl port-forward svc/monitoring-kube-prometheus-prometheus 9090
```

```bash
kubectl port-forward svc/monitoring-grafana 3000:80
```

## Dashboards
https://github.com/grafana/helm-charts/blob/main/charts/grafana/README.md#sidecar-for-dashboards

Dashboards in separate folder:
- More readable
- No escaping needed for double curly brackets 
- IDE still can support to JSON validation
- BEAWARE: size of a ConfigMap is limited to 1MB (There's a 1MB limit from the etcd side which is where Kubernetes stores its objects.)

## Install

```bash
helm install monitoring ./deploy/kubernetes/monitoring -n monitoring --create-namespace
```
