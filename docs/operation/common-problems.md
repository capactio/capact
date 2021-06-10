# Common problems

This document describes how to identify and resolve common Capact problems that might occur.

## Table of contents

<!-- toc -->

- [Action](#action)
  * [Action does not have status](#action-does-not-have-status)
  * [Action stuck in the `BeingRendered` phase](#action-stuck-in-the-beingrendered-phase)
  * [Action in the `Failed` phase](#action-in-the-failed-phase)
  * [Clean up Action execution pods](#clean-up-action-execution-pods)
  * [Wrong Implementation was selected](#wrong-implementation-was-selected)
- [Unreachable Gateway](#unreachable-gateway)

<!-- tocstop -->

## Action

In this section, you can find common [Action](./../terminology.md#action) failures that might occur.

### Action does not have status

Symptoms:

- [Action status](diagnostics.md#checking-action-status) is empty.

Debugging steps:

- [Check if Engine is up and running](diagnostics.md#engine-health).

- [Check the Engine logs](diagnostics.md#engine-logs). You can grep logs using Action name. This will narrow-down the number of log entries. During the initial process, Engine tries to update Action. The common problem can be that the Engine has wrong/missing RBAC.

### Action stuck in the `BeingRendered` phase

Rendering more complex workflow may take a few minutes. An Action in `BeingRendered` for more than 15 minutes may mean that it is stuck.

Symptoms:

- An Action was created more than 15 minutes ago. To check the **AGE** column, run:
  
  ```bash
  kubectl get actions.core.capact.io ${ACTION_NAME} -n {ACTION_NAMESPACE}
  ```

Debugging steps:

- [Check if Engine is up and running](diagnostics.md#engine-health).

- [Check the Engine logs](diagnostics.md#engine-logs). You can grep logs using Action name. This will narrow-down number of log entries. During the render process, manifests are downloaded from the Public OCH. The common problem can be that the Public OCH is unreachable for Engine. Check [Unreachable Gateway](#unreachable-gateway) section to resolve this issue.

### Action in the `Failed` phase

The action may fail for a variety of reasons. First what you need to do is to check the status message.

Debugging steps:

- [Check the Action status message](diagnostics.md#checking-action-status-message). If status message contains: `while fetching latest Interface revision string: cannot find the latest revision for Interface "cap.interfac.db.install" (giving up - exceeded 15 retries)`:

    - [Ensure that Public OCH is populated and manifests can be fetched](diagnostics.md#checking-if-public-och-is-populated).
	- Ensure that **ActionRef** is not misspelled.

- [Check the Engine logs](diagnostics.md#engine-logs). You can grep logs using Action name. This will narrow-down the number of log entries. The common problem can be that the Engine doesn't have proper permission to schedule Action execution, e.g. cannot create ServiceAccount, Secret, Argo Workflow. Ensure that the `k8s-engine-role` ClusterRole in the `capact-system` Namespace has all necessary permissions.

- [Check the Action execution](diagnostics.md#checking-action-execution-status).

### Clean up Action execution pods

After Action execution there are a lot of Pods with name pattern `{ACTION_NAME}-{RANDOM_10_DIGITS}` in the`Completed` state.

```bash
NAME                READY   STATUS      RESTARTS   AGE
mattermost-1602179194     0/2     Completed   0          14d
mattermost-2270774275     0/2     Completed   0          14d
mattermost-823541112      0/2     Completed   0          14d
mattermost-470211537      0/2     Completed   0          14d
mattermost-1030672350     0/2     Completed   0          14d
mattermost-147207013      0/2     Completed   0          14d
mattermost-2768336525     0/2     Completed   0          14d
mattermost-3634435893     0/2     Completed   0          14d
mattermost-4236050029     0/2     Completed   0          14d
mattermost-2282111071     0/2     Completed   0          14d
mattermost-3762917690     0/2     Completed   0          14d
mattermost-4129897782     0/2     Completed   0          14d
mattermost-1307838837     0/2     Completed   0          14d
mattermost-2309417707     0/2     Completed   0          14d
mattermost-1619688498-1   1/1     Running     0          12d
mattermost-1619688498-0   1/1     Running     0          12d
```

Those Pods were created by Argo Workflow and each of them represent executed Action step e.g. create a database, create user in the database etc. For failed Actions they are useful to debug the root cause of an error. For successfully execute Action you can remove them. To remove only Argo Workflow Pods, run:

```bash
kubectl delete workflows.argoproj.io {ACTION_NAME} -n {ACTION_NAMESPACE}
```

To remove Action and all resources associated with it (Argo Workflow Pods, ServiceAccount, user input data etc.), run:

```bash
capact action delete {ACTION_NAME} -n {ACTION_NAMESPACE}
```

### Wrong Implementation was selected

Actions may define theirs dependencies via Interfaces. Depending on cluster Policy configuration, every time user runs an Action, a different Implementation may be picked for a given Interface. 

Symptoms:

- [Rendered Action workflow](diagnostics.md#checking-rendered-action-workflow) contains Implementation which should not be used. 

- Executed Action create resources in the unexpected destination. For example, deployed PostgreSQL on a cluster instead of provisioning RDS instance on AWS side.

Debugging steps:

- [Check if proper policy exists and has proper configuration](diagnostics.md#checking-cluster-policy). Read the [Policy configuration](../feature/policy-configuration.md) document to get familiar with the syntax and available set of features.

- If you use cloud solutions, such as GCP or AWS, you need to specify TypeInstance ID in the cluster Policy. This TypeInstance must hold a subscription which allows to provision a given service on the hyperscaler side. If TypeInstance doesn't exist, Engine will ignore this configuration. [Check if TypeInstance with a given ID exists](diagnostics.md#checking-if-typeinstance-exists) 

## Unreachable Gateway

Gateway aggregates GraphQL APIs from the Capact Engine, Public OCH, and Local OCH. If one of the aggregated component is not working properly, Gateway is not working either.

Symptoms:

- Gateway responds with the `502` status code.

- [Gateway logs](diagnostics.md#gateway-logs) contain a message similar to: `while introspecting GraphQL schemas: while introspecting schemas with retry: while introspecting schemas: invalid character 'l' looking for beginning of value`.

- [Gateway Pod is frequently restarting](diagnostics.md#gateway-health).

Debugging steps:

- [Restart Gateway](diagnostics.md#pod-restart). For component name use `gateway`.

- [Check if OCH Public is up and running](diagnostics.md#public-och-health).

- [Check if OCH Public has in logs](diagnostics.md#public-och-logs) information that it was started. It should contain a message similar to: `INFO   GraphQL API is listening   {"endpoint":"http://:8080/graphql"}`.

- [Check if OCH Local is up and running](diagnostics.md#local-och-health).

- [Check if OCH Local has in logs](diagnostics.md#local-och-logs) information that it was started. It should contain a message similar to: `INFO   GraphQL API is listening   {"endpoint":"http://:8080/graphql"}`.

- [Check if Engine is up and running](diagnostics.md#engine-health).

- [Check if Engine has in logs](diagnostics.md#engine-logs) information that it was started. It should contain a message similar to: `engine-graphql   httputil/server.go:47  Starting HTTP server   {"server": "graphql", "addr": ":8080"}`.
