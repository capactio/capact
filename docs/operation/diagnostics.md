# Basic diagnostics

Here you can find the list of basic diagnostic actions that may help you look for bug causes.

## Table of contents

<!-- toc -->

- [Engine](#engine)
  * [Engine health](#engine-health)
  * [Engine logs](#engine-logs)
  * [Check Action status](#check-action-status)
  * [Check Action execution status](#check-action-execution-status)
- [Gateway](#gateway)
  * [Gateway health](#gateway-health)
  * [Gateway logs](#gateway-logs)
- [OCH Public](#och-public)
  * [OCH Public health](#och-public-health)
  * [OCH Public logs](#och-public-logs)
  * [OCH Populator logs](#och-populator-logs)
  * [Check if Public OCH is populated](#check-if-public-och-is-populated)
- [OCH Local](#och-local)
  * [OCH Local health](#och-local-health)
  * [OCH Local logs](#och-local-logs)
- [Pod restart](#pod-restart)

<!-- tocstop -->

## Engine 

This section describes [Engine](./../e2e-architecture.md#engine) related diagnostic.

### Engine health

To check if the Engine Pods are in the `Running` state, run:

```
kubectl get pod -n capact-system -l app.kubernetes.io/name=engine
```

All the containers from Pods should be in the `Running` status. Restarts number higher than 1 may also indicate problems, e.g. not enough resource, lack of permissions, network timeouts etc.

### Engine logs

If the Engine is [healthy](#engine-health), you should be able to track any bug by checking the logs. To check the logs, run:

```
kubectl logs -n capact-system -l app.kubernetes.io/name=engine -c ctrl
```

To check the logs since a given time, use the `--since-time` flag, for example:

```
--since-time=2020-03-30T10:02:08Z
```

### Check Action status

To check the Action status, run:

```
kubectl get actions.core.capact.io ${ACTION_NAME} -n {ACTION_NAMESPACE} -ojsonpath="{.status}"
```

### Check Action execution status

An Action is executed via Argo Workflows. To check the execution status you can use either Argo CLI or Argo UI:  

- Using command line. 
  
  Install the latest [stable Argo CLI for version v2.x.x](https://github.com/argoproj/argo-workflows/releases), and run:

  ```bash
  argo get {ACTION_NAME} -n {ACTION_NAMESPACE}
  ```

- Using browser.

  By default, the Argo UI is not exposed publicly. You need to port-forward the Service to your local machine: 
  
  ```bash
  kubectl -n capact-system port-forward svc/argo-server 2746
  ```

  Navigate to [http://localhost:2746](http://localhost:2746) to open Argo UI. Argo Workflow has the same name as the executed Action.

## Gateway

This section describes [Gateway](./../e2e-architecture.md#gateway) related diagnostic.

### Gateway health

To check if the Gateway Pods are in the `Running` state, run:

```
kubectl get po -n capact-system -l app.kubernetes.io/name=gateway
```

All the containers from Pods should be in the `Running` status. Restarts number higher than 1 may also indicate problems, e.g. not enough resource, lack of permissions, network timeouts etc.

### Gateway logs

If the Gateway is [healthy](#gateway-health), you should be able to track any bug by checking the logs. To check the logs, run:

```
kubectl logs -n capact-system -l app.kubernetes.io/name=gateway -c gateway
```

To check the logs since a given time, use the `--since-time` flag, for example:

```
--since-time=2020-03-30T10:02:08Z
```

## OCH Public

This section describes Public [OCH](./../e2e-architecture.md#och) related diagnostic.

### OCH Public health

To check if the OCH Public Pods are in the `Running` state, run:

```
kubectl get po -n capact-system -l app.kubernetes.io/name=och-public
```

All the containers from Pods should be in the `Running` status. Restarts number higher than 1 may also indicate problems, e.g. not enough resource, lack of permissions, network timeouts etc.

### OCH Public logs

If the OCH Public is [healthy](#och-public-health), you should be able to track any bug by checking the logs. To check the logs, run:

```
kubectl logs -n capact-system -l app.kubernetes.io/name=och-public -c och-public
```

To check the logs since a given time, use the `--since-time` flag, for example:

```
--since-time=2020-03-30T10:02:08Z
```

### OCH Populator logs

If the OCH Public is [healthy](#och-public-health), you should be able to track any bug by checking the logs. To check the logs, run:

```
kubectl logs -n capact-system -l app.kubernetes.io/name=och-public -c och-public-populator
```

To check the logs since a given time, use the `--since-time` flag, for example:

```
--since-time=2020-03-30T10:02:08Z
```

### Check if Public OCH is populated 

- Check if [OCH Populator logs](#och-populator-logs) contain a message similar to: `{"level":"info","ts":1620895282.3582015,"caller":"register/ocf_manifests.go:107","msg":"Populated new data","duration (seconds)":235.525841306}`. It means that manifests were populated successfully. If you get an error similar to: `error: container och-public-populator is not valid for pod capact-och-public-84cc74bc66-pmkhp` it means that the Public OCH Populator is disabled. To enable it, run:

  ```bash
  helm repo add capactio https://storage.googleapis.com/capactio-awesome-charts
  helm upgrade capact capactio/capact -n capact-system --reuse-values --set och-public.populator.enabled=true
  ```

- Check if manifests can be fetched from the Public OCH. Install the latest [stable Capact CLI](https://github.com/Project-Voltron/go-voltron/releases), and run:

  ```bash
  capact login
  capact hub interfaces search
  ```

  Successful response, should look similar to:
  ```
                             PATH                             LATEST REVISION                           IMPLEMENTATIONS
  +---------------------------------------------------------+-----------------+-----------------------------------------------------------------+
    cap.interface.analytics.elasticsearch.install             0.1.0             cap.implementation.elastic.elasticsearch.install
                                                                                cap.implementation.aws.elasticsearch.provision
  +---------------------------------------------------------+-----------------+-----------------------------------------------------------------+
    cap.interface.atlassian.stack.install                     0.1.0             cap.implementation.atlassian.stack.install
                                                                                cap.implementation.atlassian.stack.install-parallel
  +---------------------------------------------------------+-----------------+-----------------------------------------------------------------+
    cap.interface.automation.concourse.change-db-password     0.1.0             cap.implementation.concourse.concourse.change-db-password
  +---------------------------------------------------------+-----------------+-----------------------------------------------------------------+
  ...
  ```

- Check if manifest source is correct, run:

  ```bash
  kubectl get deploy capact-och-public -o=jsonpath='{$.spec.template.spec.containers[?(@.name=="och-public-populator")].env[?(@.name=="MANIFESTS_PATH")].value}'
  ```
  
  Check the [go-getter](https://github.com/hashicorp/go-getter) project to understand URL format.   

## OCH Local

This section describes Local [OCH](./../e2e-architecture.md#och) related diagnostic.

### OCH Local health

To check if the OCH Local Pods are in the `Running` state, run:

```
kubectl get po -n capact-system -l app.kubernetes.io/name=och-local
```

All the containers from Pods should be in the `Running` status. Restarts number higher than 1 may also indicate problems, e.g. not enough resource, lack of permissions, network timeouts etc.

### OCH Local logs

If the OCH Local is [healthy](#och-local-health), you should be able to track any bug by checking the logs. To check the logs, run:

```
kubectl logs -n capact-system -l app.kubernetes.io/name=och-local -c och-local
```

To check the logs since a given time, use the `--since-time` flag, for example:

```
--since-time=2020-03-30T10:02:08Z
```

## Pod restart

When Pods are unhealthy, or if the operation processing is stuck, you can restart the Pod using this command:

```
kubectl delete po -n capact-system -l app.kubernetes.io/name={COMPONENT_NAME}
```
