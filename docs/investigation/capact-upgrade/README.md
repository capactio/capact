# Upgrade Capact components via Action

Created on 2021-04-01 by Mateusz Szostok ([@mszostok](https://github.com/mszostok/))

This document describes the initial investigation about upgrading Capact components via Action.

## Table of Contents

<!-- toc -->

- [Investigation](#investigation)
  * [Versioning](#versioning)
  * [How to access installation resources](#how-to-access-installation-resources)
  * [Capact upgrade Action](#capact-upgrade-action)
    + [Known issues](#known-issues)
    + [Advantages](#advantages)
    + [Disadvantages](#disadvantages)
  * [Capact Helm charts](#capact-helm-charts)
    + [Notes](#notes)
  * [CI/CD strategy](#cicd-strategy)
  * [CLI](#cli)
  * [Others](#others)
  * [Bugs](#bugs)
  * [TODO](#todo)
  * [Proposed solution](#proposed-solution)

<!-- tocstop -->

## Investigation

### Versioning

For now, we need to take into account such versions:
  - Helm chart version (**version** in `Chart.yaml`)
     - depends on final solution, we will have only single Helm chart, or need to additionally maintain and version all dependencies (argo, neo4j etc.)
  - Capact Docker images version
     - For releases should be in sync with **appVersion** from `Chart.yaml`?
     - For long-running dev cluster just use commit SHA and ignore **appVersion**?
  - CLI version
     - it does an actual upgrade, should be versioned along with the Capact?
     - which version should we use on CI/CD, just `go run ?`, or always deploy the latest CLI version and use it? 
  - Capact upgrade Action revision
    - it will be good to also be able to define it. It will allow us e.g. to introduce automated backup and restore in the near future.

### How to access installation resources

1. Helm Chart

	Options:
	1. GCS
	2. GitHub Page, via https://github.com/helm/chart-releaser-action/blob/master/action.yml
	3. Clone the `go-voltron` repository (It is not supported by Helm Runner)

	What about releases process?

1.  CRDs

	Options:
	1. GCS
	2. GitHub raw object
	3. Clone the `go-voltron` repository (It is not supported by Helm Runner)

### Capact upgrade Action

1. Interface has multiple type instance for each dependency. Is it ok?
	
	1. How to handle “optional installation”?

    <details><summary>Interface input</summary>

    ```yaml
    spec:
      input:
        typeInstances:
          capact-config:
            typeRef:
              path: cap.type.capactio.capact.config
              revision: 0.1.0
            verbs: ["get", "update"]
          capact-helm-release:
            typeRef:
              path: cap.type.helm.chart.release
              revision: 0.1.0
            verbs: [ "get", "update" ]
          argo-helm-release:
            typeRef:
              path: cap.type.helm.chart.release
              revision: 0.1.0
            verbs: ["get", "update"]
          ingress-nginx-helm-release:
            typeRef:
              path: cap.type.helm.chart.release
              revision: 0.1.0
            verbs: ["get", "update"]
          kubed-helm-release:
            typeRef:
              path: cap.type.helm.chart.release
              revision: 0.1.0
            verbs: ["get", "update"]
          monitoring-helm-release:
            typeRef:
              path: cap.type.helm.chart.release
              revision: 0.1.0
            verbs: ["get", "update"]
          neo4j-helm-release:
            typeRef:
              path: cap.type.helm.chart.release
              revision: 0.1.0
            verbs: ["get", "update"]
      output:
        typeInstances:
          capact-config:
            typeRef:
              path: cap.type.capactio.capact.config
              revision: 0.1.0
    ```

    </details>

    **Answer**: TBD
    
    **Reason**: TBD

1. We have only test content on the long-running cluster, how to deal with that?
	1. Copy-paste the upgrade action also to the `test/och-content` folder?
	2. Switching och-content?
	3. Other options?

    **Answer**: TBD
    
    **Reason**: TBD

1. Should we update tests and add logic which will wait until Capact upgrade is finished?

    **Answer**: TBD
    
    **Reason**: TBD    
    
1. Where we store the upgrade migration logic, in Action, or via K8s Job in Helm charts?

    **Answer**: TBD
    
    **Reason**: TBD

1. What to support in upgrade Action:

    <details><summary>Custom Interface input</summary>

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
	
	1. Where we should store overrides?
	    - directly in Action upgrade
	    - download them from repository
	    - have them always as input parameters

        Idea:
        - “Features” like `INCREASE_RESOURCE_LIMITS` can be done via CLI, and if we will migrate bash script to CLI, then we will be able to decrease LOE.
        - We can use [this plugin](https://github.com/karuppiah7890/helm-schema-gen) to generate initial schema input. We will need to only add section about versioning.

    **Answer**: TBD
    
    **Reason**: TBD
    
    > **NOTE:** Add support in jinja for from JSON to YAML (with an indent).
    
1. Should we use it also locally? If yes, how?
    1. Deploying local helm repo?

    **Answer**: TBD
    
    **Reason**: TBD

#### Known issues

- We have dependencies in different Namespaces. As a result, it doesn't work properly with Helm upgrade runner. This will be fixed by Namespace unification.
- Engine needs to produce the ClusterRoleBinding for Action. This will be fixed by Namespace unification.

#### Advantages

- Upgrade Action is executed on cluster side, no more client timeout.
- No need to add GitHub job IP address, all traffic via Gateway.

#### Disadvantages

- We need to maintain two ways of upgrade, for local development and for clusters via Action.

### Capact Helm charts

1. Have a single Capact chart with dependencies.
   
   Cons:
     1.	It is problematic as we will have "umbrella chart", so we cannot upgrade a given dependency
     1.	Cannot run concurrently, we have only one big step with Helm runner upgrade action.

   Pros:
     1. Easy to maintain, e.g. versioning.
     1. Easy to create TI as we will have only a signle one.
     1. "built-in" disable components support.
     1. Simpler TI upgrade as we need to specify only a single TI.
     1. Simpler to bump versions for our dependecies.
	 1.	We can add `values.yaml` with our own overrides.
     1.	We have our own copy, so we are independent.

1. Store them as a separate charts with one dependency.
   
   Pros:
    1.	We have our own copy, so we are independent.
	1.	We can add `values.yaml` with our own overrides.

   Cons:
    1. There is no easy way to support component disable/enable.

1. Use upstream Helm charts directly.

   Cons:
    1. Hard to specify additional values.
    1. Hard to create initial TypeInstance.

#### Notes

1.	Storing additional `values.yaml` files in Helm chart doesn't work. For example, Helm chart with `values-higher-res-limits.yaml`, when used returns such error:
	
	```bash
	helm upgrade neo4j --install --create-namespace --namespace="neo4j" -f values-higher-res-limits.yaml capactio-awesome-charts/neo4j-helm
	  Release "neo4j" does not exist. Installing it now.
	  Error: open values-higher-res-limits.yaml: no such file or directory
    ```

### CI/CD strategy

1. Build and push components Docker images.
1. Detect changes in `deploy/kubernetes/charts` dir, (we need to remember about version in Chart.yaml) (This can be improved)
    1. Execute `hack/release-charts.sh`
1. Create Action via capact CLI
    1. Use built docker images (via **override.docker.tag**) 
    1. Use the newest Helm chart versions (or we should have sth like latest/master/stable/nightly?)
    
    Example:
    ```bash
    capact upgrade --helm-chart-version @latest --override-docker-tag <commit_sha>
    ```

1. Wait until upgrade finish

### CLI

> **NOTE:** Communicates only with Gateway. It is by design so we do not need to add GitHub job IP.

1. Finds `capact-config` TypeInstance based on **TypeRef**.
1. Creates input TypeInstances based on `capact-config.uses` relation.
1. Creates input parameters from user input (via flags).
1. Generates Action upgrade name.
1. Gets the latest Helm chart based on `index.yaml` from `capactio-awesome-charts` repository.
1. Creates `cap.interface.capactio.capact.upgrade` Action.
1. Waits until Action is ready to run.
1. Executes Action.
1. (Optionally) Waits until Action is finished.

### Others

Questions:
- Add `revision` property for helm releases? In that way it is easy to check consistency between TI and `helm list`
- Add CronJob which will delete old upgrade Action from long-running cluster? or always deletes them after success upgrade?

### Bugs

- Annotate secret: 
    ```yaml
    command: [ kubectl ] 
    args: [ "annotate", "secret", "-n=argo", "argo-minio", 'kubed.appscode.com/sync=""', "--overwrite" ]
    ```
   generates `kubed.appscode.com/sync=`'"""'`. Need to be `args: [ "annotate", "secret", "-n=argo", "argo-minio", 'kubed.appscode.com/sync=', "--overwrite" ]`
- Problem with `minio` Secret that is not sync from `argo` Namespace to `voltron-system` Namespace. Will be resolved when moved to single Namespace.
- Bug with Helm upgrade when there is a new chart version (cache problem). Maybe we should have option to force Helm chart download.

### TODO

- Add support in jinja for from JSON to YAML (with an indent)
- Consider getting rid of initializer and replace with simple script
- Update CI/CD pipeline, update CI/CD documentation
- Build and publish `capact` CLI (?)
- Test on GKE dev-cluster

### Proposed solution

Versioning:
1. Capact + CLI are versioned together.
1. Use the latest CLI on CI/CD via `go build` and later `capact upgarde ...`. Release CLI only when Capact is released.
1. Have dedicated Helm charts with dependencies. 
1. Helm chart **appVersion** same as Capact version, bump only for releases.
1. Helm chart **version** always changed when chart is changed and pushed to GCS but to other repository e.g. `capactio-master-charts`.
    In that way we will not have problem with a lot of version in the official repository. Maybe we an also use commit SHA as **version** on the master branch. 
    For normal releases, we bump the **version** manually and push chart to `capactio-awesome-charts`.

Upgrade Action:  
1. Accepts all parameters which can be set via `values.yaml`.
1. Accepts multiple TI.
1. Doesn't support component disable/enable feature.
1. Our own overrides (e.g. increaseResourceLimits) are stored in CLI which knows how to build input parameters JSON for upgrade Action.
1. Stores the upgrade migration logic in Action, not in Helm charts.

OCH:
1. Copy-paste the upgrade action also to the `test/och-content` folder via bash script?

Local cluster upgrade:
1. For now not supported via Action CR. In the near future it will be good to migrate to it.
