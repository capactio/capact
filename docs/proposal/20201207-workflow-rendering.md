# Workflow rendering

Created on 2020-12-07 by Damian Czaja ([@trojan295](https://github.com/trojan295))

## Overview

This document shows how we can import existing Implementations into a new Implementation and render the final workflow in Voltron Engine.

<!-- toc -->
- [Motivation](#motivation)
  * [Goal](#goal)
  * [Non-goal](#non-goal)
- [Proposal](#proposal)
  * [How to reference a Interface/specific Implementation to be called in a Implementation workflow](#how-to-reference-a-interfacespecific-implementation-to-be-called-in-a-implementation-workflow)
  * [How to reference inputs needed and outputs provided by the called Implementation](#how-to-reference-inputs-needed-and-outputs-provided-by-the-called-implementation)
  * [How to conditionally call the Implementation](#how-to-conditionally-call-the-implementation)
  * [How to render the final Argo workflow](#how-to-render-the-final-argo-workflow)
- [Consequences](#consequences)

<!-- tocstop -->

## Motivation

In many cases a Content Creator would like to leverage a Implementation workflow, which is already available in OCH. For example - they are creating a workflow to provision Wordpress and they need an PostgreSQL database. They have already an Implementations for the `postgresql.install` Interface available in OCH and they would like to use it in their Implementation.

Content Creator would like to have a syntax to call an another Implementation or Interface. For example, they are creating a Wordpress install Implementation and need a PostgreSQL database. There is already an OCH Interface `postgres.install` defined, which handles creating a PostgresSQL database. The Content Creator would also like to define an optional TypeInstance input `postgresDatabase` and execute the `postgres.install` Interface only in case this TypeInstance was not provided.

Content Creator should be able to:
- reference a Interface/specific Implementation to be called in his Implementation workflow
- reference output TypeInstances, provided by the called Interface/Implementation
- conditionally call the Implementation

Voltron Engine must be able to:
- parse the Implementation with Interface/Implementation calls and render the final Argo workflow

### Goal

- How to reference a Interface/specific Implementation to be called in a Implementation workflow
- How to reference inputs needed and outputs provided by the called Implementation
- How to conditionally call the Implementation
- How to parse the Implementation with Interface/Implementation calls and render the final Argo workflow

### Non-goal

- How Implementation inputs are provided into the workflow
- How output TypeInstance from the Implementation and called Implementations are uploaded to OCH

## Proposal

### How to reference a Interface/specific Implementation to be called in a Implementation workflow

To reference the Implementataion/specific Implementation, which must be called the following extensions to the Argo Workflow is proposed:

- `.spec.action.args.workflow.entrypoint.templates[].steps[][].action` - defines the Interface/Implementation to be called. In case this is set, the Content Creator does not have to provide a `template`

```yaml
action:                                   # optional
  name: cap.interface.postgresql.install  # required - defines the called Interface/Implementation
  revision: 0.1.0                         # optional - defines the revision of the called Interface/Implementation
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

### How to reference inputs needed and outputs provided by the called Implementation

Argo Workflows already provides a syntax for passing inputs and getting outputs from Workflow Steps:
```yaml
workflow:
  entrypoint: main
  templates:
    - name: main
      steps:
        - - name: generate
            template: generate

        - - name: consume
            template: consume
            arguments:
              artifacts:
                - name: input
                  from: "{{steps.generate.outputs.artifacts.example}}"

    - name: generate
      outputs:
        artifacts:
          - name: example
      # [...]

    - name: consume
      inputs:
        artifacts:
          - name: input
      # [...]
```

There is no need to extend the Argo workflow syntax as we can leverage the already existing mechanism for passing inputs and getting outputs from Implementation workflows

### How to conditionally call the Implementation

Argo Workflows already provides a syntax for conditional execution:
```yaml
entrypoint: coinflip
templates:
- name: coinflip
  steps:
  - - name: flip-coin
      template: flip-coin
  - - name: heads
      template: heads
      when: "{{steps.flip-coin.outputs.result}} == heads"
    - name: tails
      template: tails
      when: "{{steps.flip-coin.outputs.result}} == tails"

- name: flip-coin
  script:
    image: python:alpine3.6
    command: [python]
    source: |
      import random
      result = "heads" if random.randint(0,1) == 0 else "tails"
      print(result)

- name: heads
  container:
    image: alpine:3.6
    command: [sh, -c]
    args: ["echo \"it was heads\""]

- name: tails
  container:
    image: alpine:3.6
    command: [sh, -c]
    args: ["echo \"it was tails\""]
```

In case the condition is aplied to an Interface/Implementation call the Voltron Engine needs to fetch the referenced Implementation and render it. In this case Voltron Engine could beforehand try to evaluate the conditions, given the Action CR input Type Instances and parameters. In case an condition was able to be evaluated and it was negative, the Engine could skip including the Implementation call in the final rendered workflow. For the sake of simplicity I recommend to skip the Engine condition evaluation for now. This looks like an addon to the Engine logic and could be done at a later step.

### How to render the final Argo workflow

Above we introduced an extensions to the Argo workflow template:
- `.spec.action.args.workflow.entrypoint.templates[].steps[][].action` - defines the Interface/Implementation to be called

For rendering the final workflow we use the feature of Argo Workflows to execute nested Workflow. As every Implementation defines a Argo workflow we can reduce the OCH Implementation call to an Argo Workflow nested workflow.

The proposed algorithm for including a nested workflow from an Implementation:

1. Create the `rendered workflow` from the Implementation evaluated from the Action CR
2. Find all `WorkflowSteps` in the rendered workflow, which have the `.action` field set. If none are found, then the rendering is complete. If there are some, then foreach:
   - create the `imported workflow` for the called Implementation defined in `.action.name`
   - append all templates from the `imported workflow` to the `rendered workflow`. Prefix the template names and global variables in the imported templates with a random string
   - remove the `.action` property in the `WorkflowStep`. Set the `.template` property to the entrypoint of the imported workflow
3. Repeat 2.

From the rendering point of view it does not matter, if a reference to an Interface or Implementation is provided as long as we can get a Implementation for the Interface from OCH.

The Content Creator could not be aware of the template names and global artifacts used in the called OCH Implementation. The Engine needs to handle template name and global artifact names collisions. Prefixing with a random string is used here.

#### PostgreSQL install example

We want to create a Postgres install Implementation using Helm runner.

`helm.run` Implementation does not use any syntax, which has to be rendered. From the Content Creator point of view it exposes an interface, which:
- takes an input artifacts called `helm-args`, specified in the `helm.run` Interface
- returns two output artifacts:
  - `helm-release` - specified in the `helm.run` Interface
  - `additional` - is a result of templating the template provided in `helm-args`. It's the Content Creator's choice, what will be the content of this artifact

<details><summary>helm.run Implementation</summary>

```yaml
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

Using this we can create the following `postgres.install` Implementation:

<details><summary>postgres.install Implementation</summary>

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

- We are extending the Argo workflow syntax, which is used now in Implementation
- We are using standard Argo way of passing artifacts, no special syntax is added. Content Creator must know the inputs and output names of the artifacts used in the imported actions. This is not a problem, as they are defined by the required and optional TypeInstances of the imported actions
- We must remember prefixing names in the imported actions. Workflows could generate global artifacts and we could have collision
- When creating a new Implementation we have to keep in mind that it could be called somewhere and used as a nested workflow, not only the entrypoint
- Voltron Engine needs to upload the TypeInstances not only from the main Implementation, but also from the called Implementations. This is not in scope of this proposal, but it should be possible to detect the TypeInstances during the rendering, create a list of them and add steps to upload them to OCH
