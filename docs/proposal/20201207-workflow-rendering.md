# Workflow rendering

Created on 2020-12-07 by Damian Czaja ([@trojan295](https://github.com/trojan295))

## Overview

This document describes the approach for rendering workflows in the Voltron engine.

<!-- toc -->

- [Motivation](#motivation)
  * [Goal](#goal)
  * [Non-goal](#non-goal)
- [Proposal](#proposal)
  * [How to reference and call other Interfaces and Implementations](#how-to-reference-and-call-other-interfaces-and-implementations)
  * [How to render the final Argo workflow](#how-to-render-the-final-argo-workflow)
- [Consequences](#consequences)
<!-- tocstop -->

## Motivation

In many cases you would like to leverage an already existing action in your workflow. For example - you are creating a workflow to provision Wordpress and you need an PostgreSQL database. You have already an implementations for the `postgresql.install` interface available in your OCH and you would like to use it in your workflow.

This proposal shows how we can import existing actions into a new workflow.

### Goal

- How to reference and call other Interfaces and Implementations
- How to render the final Argo workflow

### Non-goal

- How Action CR inputs are provided into the workflow
- How output TypeInstance are uploaded to OCH

## Proposal

### How to reference and call other Interfaces and Implementations

The following extensions are done to the Implementation definitions:

- `.spec.action.args.workflow.entrypoint.templates[].steps[][].action` - defines the interface/implementation to be imported into the workflow steps. In case this is set, the Content Creator does not have to provide a `template` for this workflow step, as the renderer will automatically fill it.

```yaml
action:                                   # optional
  name: cap.interface.postgresql.install  # required
  revision: 0.1.0                         # optional
  namePrefix: postgres-install            # optional
```

<details><summary>Example</summary>

```yaml
kind: Implementation
spec:
  action:
    runnerInterface: argo.run
    args:
      workflow:
        entrypoint: jira-install
        templates:
          - name: jira-install
            steps:
              - - name: install-db
                  action:
                    name: cap.interface.postgresql.install
                    prefix: postgres-install
```
</details>

### How to render the final Argo workflow

1. Fetch the root Implementation for the Action. Create the Workflow
2. Find all `WorkflowSteps` in the rendered workflow, which have the `Action` field set. If none are found the rendering is complete. If there are some, then foreach:
   - import the implementation/interface based on the `.action` property
   - create the workflow for the imported implementation
   - append all templates from the imported workflow to the rendered workflow. Prefix the template names with `.action.namePrefix` or a random string
   - remove the `.action` property in the `WorkflowStep`. Set the `.template` property to the entrypoint of the imported workflow
3. Repeat 2.

From the rendering point of view it does not matter, if a reference to an interface or implementation is provided as long as we can get a implementation for the interface from OCH.

#### PostgreSQL install example

We want to create a Postgres install implementation using Helm runner.

`helm.run` implementation does not use any syntax, which has to be rendered. From the Content Creator point of view it exposes an interface, which:
- takes an input artifacts called `helm-args`, specified in the `helm.run` interface
- returns two output artifacts:
  - `helm-release` - specified in the `helm.run` interface
  - `additional` - is a result of templating the template provided in `helm-args`. It's the Content Creator's choice, what will be the content of this artifact

<details><summary>helm.run implementation</summary>

```yaml
# helm.run implementation
  action:
    runnerInterface: argo.run
    args:
      workflow:
        entrypoint: helm
        templates:
          - name: helm
            inputs:
              artifacts:
                - name: helm-args
                  path: "/helm-args"
            outputs:
              artifacts:
                - name: helm-release
                  path: "/helm-release"
                - name: additional
                  path: "/additional"
            container:
              image: gcr.io/projectvoltron/helm-runner:de65286
              env:
                - name: RUNNER_INPUT_PATH
                  value: "{{inputs.artifacts.helm-args.path}}"
```
</details>

Using this we can create the following `postgres.install` implementation:

<details><summary>postgres.install implementation</summary>

```yaml
  action:
    runnerInterface: argo.run
    args:
      workflow:
        entrypoint: postgres-install
        templates:
          - name: postgres-install
            outputs:
              # artifact names match with the typeInstance names so they can be referenced, when creating steps for uploading to OCH
              artifacts:
                - name: postgresql
                  from: "{{steps.postgres-install.outputs.artifacts.additional}}"
                - name: helm-release
                  from: "{{steps.postgres-install.outputs.artifacts.helm-release}}"
            steps:
              - - name: create-install-config
                  action:
                    name: cap.implementation.jinja2.template # implementation of a Jinja2 templater (runner?) :D. It templates the 'template' artifact using 'values' and outputs the 'render' artifact . Used for glueing outputs-inputs between steps.
                  arguments:
                    artifacts:
                      - name: values
                        # mock cap.type.database.postgresql.install-input typeInstance artifact, which would normally be provided by Voltron engine in some prefetch step
                        raw:
                          data: |
                            superuser:
                              username: postgres
                              password: s3cr3t
                            defaultDBName: postgres
                      - name: template
                        raw:
                          data: |
                            context:
                              name: "helm-runner-example"
                              dryRun: false
                              timeout: "10m"
                              platform:
                                namespace: "default"
                                serviceAccountName: "helm-runner-example"
                            args:
                              command: "install"
                              generateName: true
                              chart:
                                name: "postgresql"
                                repo: "https://charts.bitnami.com/bitnami"
                              values:
                                image:
                                  pullPolicy: Always
                                postgresqlDatabase: {{ defaultDBName }}
                                postgresqlPassword: {{ superuser.password }}
                              output:{% raw %}
                                additional:
                                  value: |-
                                    host: "{{ template "postgresql.fullname" . }}"
                                    port: "{{ template "postgresql.port" . }}"
                                    defaultDBName: "{{ template "postgresql.database" . }}"
                                    superuser:
                                      username: "{{ template "postgresql.username" . }}"
                                      password: "{{ template "postgresql.password" . }}"{% endraw %}

              - - name: postgres-install
                  # helm.run step
                  action:
                    name: cap.implementation.runner.helm.run
                    prefix: helm-install
                  arguments:
                    artifacts:
                      - name: helm-args
                        from: "{{steps.create-install-config.outputs.artifacts.render}}"
```
</details>

This is gonna be rendered into the following workflow:

<details><summary>Rendered postgres.install Argo workflow</summary>

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: render-poc
spec:
  entrypoint: postgres-install
  templates:
  - name: postgres-install
    outputs:
      artifacts:
      - from: '{{steps.postgres-install.outputs.artifacts.additional}}'
        name: postgresql
      - from: '{{steps.postgres-install.outputs.artifacts.helm-release}}'
        name: helm-release
    steps:
    - - arguments:
          artifacts:
          - name: values
            raw:
              data: |
                superuser:
                  username: postgres
                  password: s3cr3t
                defaultDBName: postgres
          - name: template
            raw:
              data: |
                context:
                  name: "helm-runner-example"
                  dryRun: false
                  timeout: "10m"
                  platform:
                    namespace: "default"
                    serviceAccountName: "helm-runner-example"
                args:
                  command: "install"
                  generateName: true
                  chart:
                    name: "postgresql"
                    repo: "https://charts.bitnami.com/bitnami"
                  values:
                    image:
                      pullPolicy: Always
                    postgresqlDatabase: {{ defaultDBName }}
                    postgresqlPassword: {{ superuser.password }}
                  output:{% raw %}
                    directory: "/"
                    additional:
                      value: |-
                        host: "{{ template "postgresql.fullname" . }}"
                        port: "{{ template "postgresql.port" . }}"
                        defaultDBName: "{{ template "postgresql.database" . }}"
                        superuser:
                          username: "{{ template "postgresql.username" . }}"
                          password: "{{ template "postgresql.password" . }}"{% endraw %}
        name: create-install-config
        template: c5OVu5-template # .action replaced with .template
    - - arguments:
          artifacts:
          - from: '{{steps.create-install-config.outputs.artifacts.render}}'
            name: helm-args
        name: postgres-install
        template: helm-install-helm  # .action replaced with .template

  # imported from cap.implementation.jinja2.template
  - container:
      args:
      - /template.yml
      - /values.yml
      - --format=yaml
      - -o
      - /render.yml
      image: dinutac/jinja2docker
      name: ""
      resources: {}
    inputs:
      artifacts:
      - name: template
        path: /template.yml
      - name: values
        path: /values.yml
    name: c5OVu5-template
    outputs:
      artifacts:
      - name: render
        path: /render.yml

  # imported from cap.implementation.runner.helm.run
  - container:
      env:
      - name: RUNNER_INPUT_PATH
        value: '{{inputs.artifacts.helm-args.path}}'
      image: gcr.io/projectvoltron/helm-runner:de65286
      name: ""
      resources: {}
    inputs:
      artifacts:
      - name: helm-args
        path: /helm-args
    name: helm-install-helm
    outputs:
      artifacts:
      - name: helm-release
        path: /helm-release
      - name: additional
        path: /additional
```
</details>

## Consequences

- We are using standard Argo way of passing artifacts, no special syntax is added. Content Creator must know the inputs and output names of the artifacts used in the imported actions. This is not a problem, as they are defined by the required and optional TypeInstances of the imported actions
- We must remember prefixing names in the imported actions. Workflows could generate global artifacts and we could have collisions. Open question is how to fetch the TypeInstance artifacts of imported actions with changed names (if there is a need for that)
- Open point is how to handle conditional imports. Lets say we have a `jira.install` implementation, which requires a `postgresql.config` TypeInstance. You could provide it to the action or create it, if not provided. We could just let Argo handle the conditions and check, if the `postgresql.config` TypeInstance artifacts is available or not, but maybe we could determine this during rendering instead, to avoid importing unnecesary actions. On the other hand:
> Premature optimization is the root of all evil - Donald Knuth
