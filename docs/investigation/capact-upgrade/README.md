# Upgrade Capact components via Action

Created on 2021-04-01 by Mateusz Szostok ([@mszostok](https://github.com/mszostok/))

This document describes the initial investigation about upgrading Capact components via Action.

## Table of Contents

<!-- toc -->

- [Investigation](#investigation)
  * [How to access installation resources](#how-to-access-installation-resources)
  * [Capact upgrade Action](#capact-upgrade-action)
  * [Capact Helm Charts](#capact-helm-charts)
  * [Others](#others)
  * [Bugs](#bugs)
  * [TODO](#todo)

<!-- tocstop -->

## Investigation

### How to access installation resources

1.	Helm Chart

	Options:
	1.	GCS
	2.	GitHub Page
	3.	Clone the `go-voltron` repository

	What about releases process?

1.	How to access CRDs

	Options:
	1.	GCS
	2.	GitHub raw object
	3.	Clone the `go-voltron` repository

### Capact upgrade Action

1.	Interface has multiple type instance for each dependency?
	1.	How to handle “optional installation”?

1.	We have a test content, how to deal with that?
	1.	Copy-paste it there
	2.	Switch content?

1.	Update tests and wait until Capact upgrade is finished?

1.	What to support:
	- OVERRIDE_DOCKER_TAG
	- OVERRIDE_DOCKER_REPOSITORY
	- INCREASE_RESOURCE_LIMITS

	Idea:
	- “Features” like `INCREASE_RESOURCE_LIMITS` can be done via CLI, and if we will migrate bash script to CLI too, then we will be able to simplify this
	- Normally you will need to specify what you want to do.

Limitations
- Doesn't’t support component disable/enable
- Different namespaces, doesn’t work properly with helm upgrade: will be fixed by ns unification
- Engine also changed to clusterrolebinding: will be fixe by ns unification

Advantages:
- Upgrade is executed on cluster side, no more client timeout
- No need to add job IP address, all traffic via Gateway

Disadvantages:
- We need to maintain two ways of upgrade, for local development and for clusters via Action

### Capact Helm charts

1.	Have a single capact chart with dependencies
	1.	It is problematic as we will have “umbrella chart”
	2.	Easy to maintain, easy to create TI, deterministic, “built-in” disable support
	3.	Simpler TI upgrade
	4.	Simpler to bump the deps,

1.	Store them as a separate charts with one dep
	1.	We have our own copy
	2.	We can add `values.yaml` with our own overrides  

1.	Storing values in helm chart doesn’t work
	1.	e.g. values-higher-res-limits.yaml
	2.	helm upgrade neo4j --install --create-namespace --namespace="neo4j" -f values-higher-res-limits.yaml capactio-awesome-charts/neo4j-helm Release "neo4j" does not exist. Installing it now. Error: open values-higher-res-limits.yaml: no such file or directory

### Others

Questions:
- Add `revision` property for helm releases? In that way it is easy to check consistency between TI and `helm list`

### Bugs

- Annotate secret: 
    ```yaml
    command: [ kubectl ] 
    args: [ "annotate", "secret", "-n=argo", "argo-minio", 'kubed.appscode.com/sync=""', "--overwrite" ]
    ```
   generates `kubed.appscode.com/sync=`'"""'`. Need to be `args: [ "annotate", "secret", "-n=argo", "argo-minio", 'kubed.appscode.com/sync=', "--overwrite" ]`
- Problem with secret that is not sync from `argo` Namespace to `voltron-system` Namespace. Will be resolved when moved to single Namespace.
- Bug with Helm upgrade when there is a new chart version (cache problem). Maybe we should have option to force Helm chart download.

### TODO

- Add support in jinja for from JSON to YAML (with an indent)
- Consider getting rid of initializer and replace with simple script
