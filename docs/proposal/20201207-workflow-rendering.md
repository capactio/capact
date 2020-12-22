# Workflow rendering

Created on 2020-12-22 by Damian Czaja ([@trojan295](https://github.com/trojan295))

## Overview

This document shows how we can render the workflow in Voltron Engine.

<!-- toc -->
- [Workflow rendering](#workflow-rendering)
  - [Overview](#overview)
  - [Motivation](#motivation)
    - [Goal](#goal)
  - [Proposal](#proposal)
    - [How to reference a Interface to be called in a Action workflow](#how-to-reference-a-interface-to-be-called-in-a-action-workflow)
    - [How to pass input parameters to the called Interface](#how-to-pass-input-parameters-to-the-called-interface)
    - [How to reference output TypeInstances from a called Interface](#how-to-reference-output-typeinstances-from-a-called-interface)
    - [How to conditionally call a Interface](#how-to-conditionally-call-a-interface)
    - [How to define, which workflow artifacts are TypeInstances](#how-to-define-which-workflow-artifacts-are-typeinstances)
    - [How to populate input parameters to the Action workflow](#how-to-populate-input-parameters-to-the-action-workflow)
    - [How to merge called Interfaces into the Action workflow](#how-to-merge-called-interfaces-into-the-action-workflow)
  - [Consequences](#consequences)

<!-- tocstop -->

## Motivation

In many cases a Content Creator would like to leverage a Interface, which is already available in OCH. For example - they are creating a workflow to provision Wordpress and they need an PostgreSQL database. They have already an Interface `postgresql.install` Interface available in OCH and they would like to use it in their Implementation.

Voltron must have an option, to allow the Content Creator reference another Interface in his Implementation. This way a Content Creator can prepare Actions, which use the already existing platform capabilities.

Content Creator should be able to:
- reference a Interface to be called in a Action workflow,
- pass input parameters to the called Interface,
- reference output TypeInstances from a called Interface,
- conditionally call a Interface,
- define, which workflow artifacts are TypeInstances.

Voltron Engine must be able to:
- populate input parameters to the Action workflow,
- merge called Interfaces into the Action workflow.

For now, we want to base on the Argo workflow syntax and only extend it, to support our additional use cases. In the future we might revisit this and change the syntax, so it is more user friendly.

### Goal

- How to reference a Interface to be called in a Action workflow
- How to pass input parameters to the called Interface
- How to reference output TypeInstances from a called Interface
- How to conditionally call a Interface
- How to define, which workflow artifacts are TypeInstances
- How to populate input parameters to the Action workflow
- How to merge called Interfaces into the Action workflow

## Proposal

### How to reference a Interface to be called in a Action workflow

To reference the Interface, which have to be called, the following extensions to the Argo workflow is proposed:

- `.spec.action.args.workflow.entrypoint.templates[].steps[][].ocf-action` - defines the Interface to be called. It is mapped to `v1alpha1.ManifestReference`. If this is set, the Content Creator does not have to provide a `template` field.

```yaml
ocf-action:                               # optional
  path: cap.interface.postgresql.install  # required - defines the called Interface/Implementation
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
                  ocf-action:
                    name: cap.interface.postgresql.install
```
</details>

### How to pass input parameters to the called Interface

Interfaces need input parameters and the Content Creator must have a way to pass them to the Interfaces, he calls. For this, we define an input artifact named `input-parameters`, where the Content Creator can pass these parameters:

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
                  ocf-action:
                    name: cap.interface.postgresql.install
                  arguments:
                    artifacts:
                      - name: input-parameters
                        raw:
                          data: |
                            username: dbuser
                            dbName: testdb
```

If the Content Creators would like to allow the Voltron User to define parameters for called Interfaces, they must define them in the Interface, they implement, and pass it down to the Interface.

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
            inputs:
              artifacts:
                - name: input-parameters
            steps:
              - - name: install-db
                  ocf-action:
                    name: cap.interface.postgresql.install
                  arguments:
                    artifacts:
                      - name: input-parameters
                        from: "{{inputs.artifacts.input-parameters}}"
```

### How to reference output TypeInstances from a called Interface

The Content Creator should be able to reference and use a TypeInstance from a called Interface. Let's take an example:

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
                  ocf-action:
                    name: cap.interface.postgresql.install
              - - name: install-jira
                  template: install-jira
                  arguments:
                    artifacts:
                      - name: "postgresql"
                        from: "{{workflow.outputs.artifacts.postgresql}}"
```

The Interface has a output TypeInstance called `postgresql` defined. To expose it to the Content Creators, we export it as a global artifact called `postgresql`.  They can then reference it using `{{workflow.outputs.artifacts.postgresql}}`.

### How to conditionally call a Interface

Content Creators might make their Actions self-sufficient and create the dependent TypeInstances in it, or allow the Voltron User to provide existing TypeInstances for the Action. To support this case we need an option to conditionally call Interfaces.

As we base on Argo, we could use the `when` directive provided by Argo, but we decided to add our own mechanism and introduce a directive `ocf-when`. Argo conditionals are evaluated during workflow execution and we need to evaluate the conditions during render-time, to do not include unnecessary workflow steps.

Only input TypeInstances can be used in the condition syntax. If there is a input TypeInstance `postgresql` defined on the Interface you are implementing, then you can make conditions based on it. E.g. `ocf-when: postgresql == nil`.

For the actual implementation aspect, we propose to use the [Expr](https://github.com/antonmedv/expr) library to evaluate the condition expressions. It is used in Argo for the `depends` directive.
In the [rendering proof-of-concept](../investigation/workflow-rendering) [govaluate](https://github.com/Knetic/govaluate) was used, but looks no longer maintained, based on GitHub activity.

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
                  ocf-when: postgresql == nil
                  ocf-action:
                    name: cap.interface.postgresql.install
```
</details>

### How to define, which workflow artifacts are TypeInstances

Output TypeInstances are defined in on the Interface, which is being implemented in an Implementation. We need a way for the Content Creator to say, that a artifact created in the Argo workflow is an TypeInstance and is supposed to be uploaded to OCH.

TypeInstance artifacts must be global Argo artifacts, so they can be fetched from the Argo Artifact Repository. In addition to that, we define a directive `ocf-type-instances` on a workflow template.

The `ocf-output-type-instances` is a list of mappings between the output TypeInstance and Argo global artifacts:
```yaml
name: {output-type-instance-name}
from: {argo-global-artifact-reference}
```

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
            ocf-output-type-instances:
              - name: jira-config
                from: "{{workflow.outputs.artifacts.some-artifact}}"
            steps:
              - - name: jira-install # This steps creates a jira-config global artifact
```

Based on this, Voltron Engine will be able to add additional steps to upload the artifacts to OCH.

### How to populate input parameters to the Action workflow

Interfaces define input parameters to the Action, which can be defined by the Voltron User. They need to be populated into the Workflow, so the Content Creator can used them.

We propose to inject them into the workflow as a local Argo artifacts named `input-parameters`. This way the Content Creator is able to use them.

```yaml
kind: Implementation
spec:
  action:
    runnerInterface: argo.run
    args:
      workflow:
        entrypoint: postgres-install
        templates:
          - name: postgres-install
            inputs:
              artifacts:
                - name: input-parameters
```

During the rendering phase, Voltron Engine will inject the `input-parameters` artifact as raw data:

```yaml
  arguments:
    artifacts:
    - name: input-parameters
      raw:
        data: |
          username: dbuser
          dbName: testdb
  entrypoint: postgres-install
  templates:
    - name: postgres-install
      inputs:
        artifacts:
          - name: input-parameters
```

### How to merge called Interfaces into the Action workflow

For rendering the final workflow we use the feature of Argo workflows to execute nested workflows. As every Implementation defines a Argo workflow we can reduce the Interface calls to an nested Argo workflow.

To avoid name collisions on the workflow templates and global artifacts, we propose to prefix the template and global artifact names with a `{template.name-step.name}` prefix, based on the step, where it was called. For example, if the Interface was called in template `jira-install` and step `helm-run`, then the prefix would be `jira-install-helm-run`.

The proposed algorithm for including a nested workflow from an called Interface is described [here](../investigation/workflow-rendering/README.md).

## Consequences

- Add `ocf-output-type-instances`, `ocf-when`, `ocf-action` directives to the workflow syntax
