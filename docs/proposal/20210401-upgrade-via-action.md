# Upgrade Capact components via Action

Created on 2021-04-01 by Mateusz Szostok ([@mszostok](https://github.com/mszostok/))

## Overview

This document describes how Capact components can be updated via dedicated Capact [Capability](https://capact.io/docs/terminology#capability).

<!-- toc -->

- [Motivation](#motivation)
- [Goal](#goal)
- [Proposal - Selected solution](#proposal---selected-solution)
  * [Capact installation](#capact-installation)
  * [Access installation resources](#access-installation-resources)
  * [Capact Helm charts](#capact-helm-charts)
  * [Versioning](#versioning)
  * [Capact upgrade Action](#capact-upgrade-action)
    + [Known issues](#known-issues)
    + [Advantages](#advantages)
    + [Disadvantages](#disadvantages)
    + [Notes](#notes)
  * [CI/CD strategy](#cicd-strategy)
  * [CLI](#cli)
- [Alternatives](#alternatives)
  * [Access installation resources](#access-installation-resources-1)
  * [Capact upgrade Action](#capact-upgrade-action-1)
  * [Capact Helm charts](#capact-helm-charts-1)

<!-- tocstop -->

## Motivation

To simplify the upgrade process and benefit from Capact features, we want to create and use dedicated [Interface](../../ocf-spec/0.0.1/README.md#interface) and [Implementation](../../ocf-spec/0.0.1/README.md#implementation) manifests.

## Goal

Prepare the zero-downtime Capact Upgrade process via dedicated Action.

## Proposal - Selected solution

The sections below describe the solutions selected during the engineering team meeting which happened on 2021-04-02. Attendees in alphabetical order by last name: 
- [Paweł Kosiec](https://github.com/pkosiec)
- [Łukasz Oleś](https://github.com/lukaszo)
- [Mateusz Szostok](https://github.com/mszostok)

### Capact installation

Capact is installed via Helm command. We will add a dedicated `post-install` Job to execute the `populator` binary to upload initial TypeInstances which describe Capact installation. 

### Access installation resources

Helm charts and CRDs are published to the GCS bucket.

### Capact Helm charts

External Helm charts are stored as separate charts with one dependency. As an example, take the [`argo`](../../deploy/kubernetes/charts/argo) chart where the `Chart.yaml` is:

```yaml
apiVersion: v2
name: argo
description: Argo chart for Kubernetes

type: application

version: 0.2.0

dependencies:
  - name: argo
    version: "0.16.7"
    repository: https://argoproj.github.io/argo-helm
```

Pros:
  - We have our own Helm chart copy.
  - We can bundle our own overrides via `values.yaml`.
  - We can upgrade each dependency independently.
  - We can run installation and upgrade concurrently.

Cons:
  - Currently, there is no easy way to support the component disable/enable feature.

### Versioning

| Property                                 | Versioning strategy                                                                                                 |
|------------------------------------------|---------------------------------------------------------------------------------------------------------------------|
| **version** from `Chart.yaml`            | It is the same for all Helm charts (Capact and all dependencies). It is changed manually for each Capact release. |
| **appVersion** from `Chart.yaml`         | It is the same as **version** from `Chart.yaml`                                                                     |
| **DOCKER_TAG** for Capact Docker images | It is the same as **version** from `Chart.yaml`                                                                     |
| CLI version                          | It is the same as **version** from `Chart.yaml`                                                                     |
| **revision** for upgrade Action          | It is changed manually and independent of the Capact version. CLI uses the latest one.                             |

### Capact upgrade Action

This section describes final agreements for the upgrade [Action](https://capact.io/docs/terminology#action).

1. Interface requires multiple input TypeInstance which describe Capact and all its dependencies. 

    <details><summary>Interface input TypeInstances</summary>

    ```yaml
    spec:
      input:
        typeInstances:
          capact-config:
            typeRef:
              path: cap.type.capactio.capact.config
              revision: 0.1.0
            verbs: [ "get", "update" ]
          capact-helm-release:
            typeRef:
              path: cap.type.helm.chart.release
              revision: 0.1.0
            verbs: [ "get", "update" ]
          argo-helm-release:
            typeRef:
              path: cap.type.helm.chart.release
              revision: 0.1.0
            verbs: [ "get", "update" ]
          ingress-nginx-helm-release:
            typeRef:
              path: cap.type.helm.chart.release
              revision: 0.1.0
            verbs: [ "get", "update" ]
          kubed-helm-release:
            typeRef:
              path: cap.type.helm.chart.release
              revision: 0.1.0
            verbs: [ "get", "update" ]
          monitoring-helm-release:
            typeRef:
              path: cap.type.helm.chart.release
              revision: 0.1.0
            verbs: [ "get", "update" ]
          neo4j-helm-release:
            typeRef:
              path: cap.type.helm.chart.release
              revision: 0.1.0
            verbs: [ "get", "update" ]
    ```

    </details>

1. Interface supports limited input parameters. 

    <details><summary>Interface input parameters</summary>

    ```yaml
    spec:
      input:
        input-parameters:
          jsonSchema:
            value: |-
              {
                "$schema": "http://json-schema.org/draft-07/schema",
                "examples": [
                  {
                    "version": "0.1.0",
                    "increaseResourceLimit": true,
                    "override": {
                        "docker": {
                          "repository": "docker.io/capact",
                          "tag": "latest"
                        }
                    }
                  }
                ],
                "properties": {
                  "version": {
                    "type": "string"
                  },
                  "increaseResourceLimits": {
                    "type": "boolean"
                  },
                  "override": {
                    "type": "object",
                    "properties": {
                      "docker": {
                        "type": "object",
                        "properties": {
                          "repository": {
                            "type": "string"
                          },
                          "tag": {
                            "type": "string"
                          }
                        }
                      }
                    }
                  }
                }
              }
    ```

    </details>

    In the near future, we will implement a generic solution. For more info, check **Interface supports generic input parameters** from the [Access installation resources](#access-installation-resources-1) section.

1. The long-running cluster is configured with only test manifests, as we do not have federation support yet. As a result, we will not have access to the upgrade Action manifests. To fix that problem, we decided to merge `test/hub-manifests` into `hub-manifests`. Each manifest will be defined under the `validation` node. Additionally, we will give an option for others to have out-of-the-box manifests, which they can use for their own validation process.

1. Add logic to block [e2e tests](../../test) until Capact upgrade is finished.

1. If necessary, the upgrade migration logic should be defined directly in Action.

1. The upgrade Action is used only for long-running clusters. For the local development cluster, we still use [`./dev-cluster-update.sh`](../../hack/dev-cluster-update.sh). In the near future, this script will be replaced with the upgrade Action.

1. After you have successfully executed the upgrade Action, delete the Action CR.

#### Known issues

- Engine needs to produce the ClusterRoleBinding for Action. This will be fixed by Namespace unification.
- Problem with the `minio` Secret that is not synced from the `argo` Namespace to the `capact-system` Namespace. This will be fixed by Namespace unification.
- Sometimes Helm upgrade runner doesn't work when there is a new Helm chart version. This is due to the cache mechanism. We should have an option to force Helm chart download.

#### Advantages

- Upgrade Action is executed on cluster side, no more client network timeouts.
- No need to add the GitHub job IP address for the upgrade CI pipeline as all traffic goes via Gateway.

#### Disadvantages

- We need to maintain two ways of upgrade, for local development and for clusters via Action.

#### Notes

Applying additional `values.yaml` files directly from Helm chart repository doesn't work. For example, when you use the Helm chart with `values-higher-res-limits.yaml`, it returns this error:
	
	```bash
    helm upgrade neo4j --install --create-namespace --namespace="neo4j" -f values-higher-res-limits.yaml capact-awesome-charts/neo4j-helm
	  Release "neo4j" does not exist. Installing it now.
	  Error: open values-higher-res-limits.yaml: no such file or directory
    ```

### CI/CD strategy

1. Build and push components Docker images.
1. Detect changes in the `deploy/kubernetes/charts` directory.
    1. Change **version** in `Chart.yaml` to the current commit SHA (same as **DOCKER_TAG**).
    1. Execute `hack/release-charts.sh`.
    
1. Create the upgrade Action via CLI.
    1. Use the latest CLI on CI/CD via `go build`.
    1. Use the built Docker images via **override.docker.tag**.
    1. Use the newest Helm chart versions based on the **created** timestamp.
    
    Example:
    ```bash
    capact upgrade --version @latest  --override-docker-tag <commit_sha>
    ```

1. Wait until upgrade is finished.
1. After the upgrade Action succeeded, execute `capact action delete foo`. 

### CLI

> **NOTE:** CLI communicates only with the Gateway. It is by design so that we don't need to add the GitHub upgrade job IP address.

The CLI executes the upgrade Action in the following way:

1. Finds the `capact-config` TypeInstance based on the **typeRef** property.
1. Creates input TypeInstances based on the `capact-config.uses` relation.
1. Creates input parameters from user input (via CLI flags).
1. Generates the Action upgrade name.
1. Gets the latest Helm chart based on `index.yaml` from the `capact-latest-charts` repository.
1. Creates the `cap.interface.capactio.capact.upgrade` Action.
1. Waits until the Action is ready to run.
1. Executes the Action.
1. (Optionally) Waits until the Action is finished.

## Alternatives

### Access installation resources

- Helm Chart:
  - Use a GitHub Page and dedicated [chart-releaser-action](https://github.com/helm/chart-releaser-action/blob/master/action.yml), or
  - Clone the `capact` repository. Referencing locally available Helm chart is not supported by Helm Runner.

- CRDs:
  - Use the GitHub raw object, or
  - Clone the `capact` repository. Referencing locally available Helm chart is not supported by Helm Runner.

### Capact upgrade Action

Interface supports generic input parameters.

<details><summary>Generic Interface input</summary>

```yaml
spec:
  input:
    input-parameters:
      jsonSchema:
        value: |-
          {
            "$schema": "http://json-schema.org/schema#",
            "type": "object",
            "properties": {
              "global": {
                  "type": "object",
                  "properties": {
                      "containerRegistry": {
                          "type": "object",
                          "properties": {
                              "overrideTag": {
                                  "type": "string"
                              },
                              "path": {
                                  "type": "string"
                              }
                          }
                      },
                      "database": {
                          "type": "object",
                          "properties": {
                              "endpoint": {
                                  "type": "string"
                              },
                              "password": {
                                  "type": "string"
                              },
                              "username": {
                                  "type": "string"
                              }
                          }
                      },
                      "domainName": {
                          "type": "string"
                      },
                      "gateway": {
                          "type": "object",
                          "properties": {
                              "auth": {
                                  "type": "object",
                                  "properties": {
                                      "password": {
                                          "type": "string"
                                      },
                                      "username": {
                                          "type": "string"
                                      }
                                  }
                              }
                          }
                      }
                  }
              },
              "integrationTest": {
                  "type": "object",
                  "properties": {
                      "expectedNumberOfRunningPods": {
                          "type": "integer"
                      },
                      "image": {
                          "type": "object",
                          "properties": {
                              "name": {
                                  "type": "string"
                              },
                              "pullPolicy": {
                                  "type": "string"
                              }
                          }
                      }
                  }
              },
              "postInstallTypeInstanceJob": {
                  "type": "object",
                  "properties": {
                      "image": {
                          "type": "object",
                          "properties": {
                              "name": {
                                  "type": "string"
                              },
                              "pullPolicy": {
                                  "type": "string"
                              }
                          }
                      }
                  }
              }
            }
          }
```

</details>

Features like increasing resource limits can be handled via CLI, and if we migrate the bash script to CLI, we will be able to decrease the LOE as we will have a single source of truth for configuration parameters. We can use the [`helm-schema-gen` plugin](https://github.com/karuppiah7890/helm-schema-gen) to generate the initial JSONSchema input files for our Helm charts. To support this way of configuration in Implementations, we need to add the function to translate JSON to YAML in Jinja Runner.
    
### Capact Helm charts

- Have a single Capact chart with dependencies.
  
  Cons:
    - It is problematic as we will have "umbrella chart", so we cannot upgrade a given dependency independently.
    - Helm install/upgrade cannot be executed concurrently. We have only one big step with Helm runner upgrade action.

  Pros:
    - Easy to maintain. For example, versioning Helm charts is easier as we need to maintain single **appVersion** and **version** properties.
    - Easy to create TypeInstances as we will have only a single one Helm release.
    - We have a built-in enable/disable components support via Helm dependency [**conditions**](https://helm.sh/docs/chart_best_practices/dependencies/#conditions-and-tags).
    - Easier Action upgrade Interface, as we need to specify only two TypeInstances.
    - We can add `values.yaml` with our own overrides.
    - We have our own copy, so we are independent.

- Use upstream Helm charts directly.

  Cons:
   - Hard to specify additional values.
   - Hard to create initial TypeInstance as the Helm repository URL will be different for each Helm chart.
