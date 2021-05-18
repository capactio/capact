# Policy configuration

## Table of Contents

<!-- toc -->

- [Introduction](#introduction)
- [Syntax](#syntax)
  * [Definition of rules for Interface](#definition-of-rules-for-interface)
  * [Selecting Implementations](#selecting-implementations)
  * [TypeInstance injection](#typeinstance-injection)
  * [Example](#example)
- [Configuration](#configuration)
  * [View current Policy](#view-current-policy)
  * [Modify Policy](#modify-policy)
  * [Reloading policy by Engine](#reloading-policy-by-engine)

<!-- tocstop -->

## Introduction

The key Capact feature is dependencies interchangeability. Applications define theirs dependencies by using Interfaces. Depending on Cluster Admin configuration, every time User runs Action, a different Implementation may be picked for a given Interface.

The Cluster Admin preferences are set via Policy. Currently, there is a single, cluster-wide Policy. This document describes the functionality.

## Syntax

Policy is defined in a form of YAML file. It contains two main features:
- selecting Implementations based on their constraints,
- injecting given TypeInstance for Implementation with a set of constraints.

### Definition of rules for Interface

You can specify which Implementations should be selected for a given Interface using `interface` property. There are three different Interface selectors:

- Interface with exact revision:

    ```yaml
    rules:
    - interface:
        path: cap.interface.database.postgresql.install
        revision: 0.1.0 # exact revision
      oneOf:
        - implementationConstraints:
            # (...)
    ```

- Interface with any revision:

    ```yaml
    rules:
    - interface:
        path: cap.interface.database.postgresql.install
        # any revision
      oneOf:
        - implementationConstraints:
            # (...)
    ```
- any Interface, using `cap.*` as an Interface path:
    
    ```yaml
    rules:
    - interface:
        path: cap.* # any Interface
      oneOf:
        - implementationConstraints:
            # (...)
    ```

Engine will search for rules for a given Interface in the same order as specified in the list above. If an entry for a given Interface is found, then Engine uses it to fetch Implementations from OCH.

For every Interface, Cluster Admin can set the order of selected Implementations, based on theirs constraints. The order of the list is important, as it is taken into account by Engine during queries to OCH. Engine iterates over list of `oneOf` items until it finds at least one Implementation satisfying the Implementation constraints.

### Selecting Implementations

You can select Implementations based on the following Implementation constraints:

- `path`, which specifies the exact path for the Implementation. If path found, then **any** revision of the Implementation is used. 

    ```yaml
    - interface:
        path: cap.interface.database.postgresql.install
      oneOf:
        - implementationConstraints:
            path: "cap.implementation.bitnami.postgresql.install" # any revision can be used
    ```

- `attributes`, which specifies which Attributes a given Implementation must contain.

    ```yaml
    - interface:
        path: cap.interface.database.postgresql.install
      oneOf:
        - implementationConstraints:
            attributes:
              - path: "cap.attribute.cloud.provider.gcp"
                revision: "0.1.0"
              - path: "cap.attribute.workload.stateful"
                # any revision
    ```

- `requires`, which specifies which Type references should be included in the `spec.requires` field for the Implementation. 

    ```yaml
    - interface:
        path: cap.interface.database.postgresql.install
      oneOf:
        - implementationConstraints:
            requires:
              - path: "cap.core.type.platform.kubernetes" # any revision
              - path: "cap.type.gcp.auth.service-account"
                revision: "0.1.0" # exact revision 
    ```

- Empty constraints, which means any Implementation for a given Interface.
    
    ```yaml
    - interface:
        path: cap.interface.database.postgresql.install
      oneOf:
        - implementationConstraints: {} # any Implementation that implements the Interface
    ```

You can also deny all Implementations for a given Interface with the following syntax:

```yaml
- interface:
    path: cap.interface.database.postgresql.install
  oneOf: [] # deny all Implementations for a given Interface
```

### TypeInstance injection

Along with Implementation constraints, Cluster Admin may configure TypeInstances, which are downloaded and injected in the Implementation workflow. For example:

```yaml
rules:
  - interface:
      path: cap.interface.database.postgresql.install 
    oneOf:
      - implementationConstraints:
          requires:
           - path: "cap.type.gcp.auth.service-account"
        injectTypeInstances:
          - id: 9038dcdc-e959-41c4-a690-d8ebf929ac0c
            typeRef:
              path: "cap.type.gcp.auth.service-account"
              revision: "0.1.0"
```

The rule defines that Engine should select Implementation, which requires GCP Service Account TypeInstance. To inject the TypeInstance in a proper place, the Implementation must define `alias` for a given requirement:

```yaml
  requires:
    cap.type.gcp.auth:
      allOf:
        - name: service-account
          alias: gcp-sa # required for TypeInstance injection based on Policy
          revision: 0.1.0
```

If the `alias` property is defined for an item from `requires` section, Engine injects a workflow step which downloads a given TypeInstance by ID and outputs it under the `alias`. For this example, in the Implementation workflow, the TypeInstance value is available under `{{workflow.outputs.artifacts.gcp-sa}}`.

Even if the Implementation satisfies the constraints, and the `alias` is not defined or `injectTypeInstances[].typeRef` cannot be found in the `requires` section, the TypeInstance is not injected in workflow. In this case Engine doesn't return an error.

### Example

The following YAML snippet presents full Policy example with additional comments:

```yaml
rules: # Configures the following behavior for Engine during rendering Action
 - interface: # Rules for Interface with exact path in exact revision
     path: "cap.interface.database.postgresql.install"
     revision: "0.1.0"   
   oneOf: # Engine follows the order of the Implementation selection,
          # finishing when at least one matching Implementation is found
     - implementationConstraints: # In first place, find and use an Implementation which:
         attributes: # contains the following Attributes:
           - path: "cap.attribute.cloud.provider.gcp"
             revision: "0.1.0" # in exact revision
         requires: # AND has the following Type references defined in the `spec.requires` property:
           - path: "cap.type.gcp.auth.service-account"
             # in any revision
       injectTypeInstances: # For such Implementation, inject the following TypeInstances: 
         - id: 9038dcdc-e959-41c4-a690-d8ebf929ac0c
           typeRef: # Find the alias of the Type reference in `spec.requires` property.
                    # If it is defined, inject the TypeInstance with ID `9038dcdc-e959-41c4-a690-d8ebf929ac0c` under this alias.
             path: "cap.type.gcp.auth.service-account"
             revision: "0.1.0"
             
     - implementationConstraints: # In second place find and select Implementation which:
         attributes: # contains the following attributes
          - path: cap.attribute.cloud.provider.aws
            # in any revision
            
     - implementationConstraints: # In third place, find and select Implementation which:
         path: "cap.implementation.bitnami.postgresql.install" # has exact path
         
      # If not found any of such Implementations defined in `oneOf`, return error.
   
  - interface:
      path: "cap.*" # For any other Interface
      # (looked up in third place, if there is no entry under `rules` for a given Interface `path:revision` or `path`)
    oneOf: # Engine follows the order of the Implementation selection,
           # finishing when at least one matching Implementation is found
      - implementationConstraints: # In first place, select Implementation which:
          requires: # has the following Type references defined in the `spec.requires` property:
            - path: "cap.core.type.platform.kubernetes"
              # in any revision

      - implementationConstraints: {} # If not found, fallback to any Implementation which has requirements that current system satisfies.

      # If not found any of such Implementations defined in `oneOf`, return error. 
```

## Configuration

You can view and update the Policy using dedicated commands in Capact CLI. You can also modify Policy during Capact installation or upgrade. This section describes all Policy configuration options.

### View current Policy

To view current Policy rules, use the following command:

```bash
capact policy get
```

You can use additional flags to configure the command behavior, such as output format. Run `capact policy get -h` to see all available flags.

### Modify Policy using CLI

> **NOTE:** If you update Policy using CLI, Policy rules will be restored to default every time you upgrade Capact installation. To avoid such scenario, update the Policy during Capact installation or upgrade. To learn how to do it, read the section [Modify Policy with Capact installation/upgrade overrides](#modify-policy-with-capact-installationupgrade-overrides).

To update Policy interactively using CLI, run:

```yaml
capact policy update
```

You can also update the Policy from YAML file, using command:

```yaml
capact policy update --from-file={path}
```

To get familiar with an example content of the file, see the [Example](#example) section.

### Modify Policy with Capact installation/upgrade overrides

1. Prepare a `cluster-policy.overrides.yaml` file with the following content:

    ```yaml
    engine:
        clusterPolicyRules:
            # Your rules here, for example:
            - interface:
                path: "cap.*"
              oneOf: 
                - implementationConstraints: # Prefer Implementations which require Kubernetes TypeInstance
                    requires:
                        - path: "cap.core.type.platform.kubernetes"
                - implementationConstraints: { } # If there are no such Kubernetes Implementations, take anything
    ```

    To know how to define the rules, see the [Syntax](#syntax) section of this document.

2. Pass the `cluster-policy.overrides.yaml` as Helm chart values override with the `-f /path/to/cluster-policy.overrides.yaml` parameter.

   1. During Capact installation:
    
   Follow the [Kubernetes installation guide](../deploy/kubernetes/README.md). While installing Capact Helm chart, provide additional overrides:
   
   ```bash
   helm install capact ./charts/capact -n capact-system -f /path/to/cluster-policy.overrides.yaml
   ```

   1. During Capact chart upgrade:

   ```bash
   helm upgrade capact ./charts/capact -n capact-system --reuse-values -f /path/to/cluster-policy.overrides.yaml
   ```

To read more about Capact installation and upgrade, see the [`README.md`](../deploy/kubernetes/charts/argo/charts/argo/README.md) document of the Capact deployment.

### Reloading policy by Engine

Once you update the ConfigMap with Policy, the Engine will reload it instantly, even for Action which are being rendered. In some cases, it may cause rendering error for a given Action. Even though the Engine will retry rendering until it reaches a configured limit of retries, it is recommended to not update the Policy while some Actions are rendering.
