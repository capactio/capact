# Voltron Metrics

To help diagnose issues the Voltron components expose a set of [Prometheus metrics](https://prometheus.io/) and [Grafana dashboards](https://grafana.com/).

We provide an optional monitoring chart [monitoring](../deploy/kubernetes/charts/monitoring) which uses under the hood the [kube-prometheus stack](https://github.com/prometheus-operator/kube-prometheus) Helm chart. The Voltron components are integrated with the provided monitoring solution. If you want to use your own monitoring stack, take into account that:

- Metrics are exposed using ServiceMonitor with `voltron.dev/scrape-metrics: "true"` label.
- Grafana dashboards are exposed via ConfigMap with `grafana_dashboard: "1"` label.

More information about exposing and adding Grafana dashboard can be found [here](./development.md#instrumentation) 
