# Workflow rendering

Created on 2020-12-22 by Damian Czaja ([@trojan295](https://github.com/trojan295))

## Overview

This document shows how we can render the workflow in Voltron Engine.

<!-- toc -->
- [Workflow rendering](#workflow-rendering)
  - [Overview](#overview)
  - [Motivation](#motivation)
    - [Goals](#goals)
  - [Proposal](#proposal)
    - [How to reference an Interface to be called in an Action workflow](#how-to-reference-an-interface-to-be-called-in-an-action-workflow)
    - [How to use and pass user input parameters to Interfaces](#how-to-use-and-pass-user-input-parameters-to-interfaces)
    - [How to reference output TypeInstances from a called Interface](#how-to-reference-output-typeinstances-from-a-called-interface)
    - [How to conditionally call an Interface](#how-to-conditionally-call-an-interface)
    - [How to define, which workflow artifacts are TypeInstances](#how-to-define-which-workflow-artifacts-are-typeinstances)
    - [How to merge called Interfaces into the Action workflow](#how-to-merge-called-interfaces-into-the-action-workflow)
  - [Example manifests with new directives](#example-manifests-with-new-directives)
  - [Consequences](#consequences)

<!-- tocstop -->

## Motivation

Voltron must, bases on the available OCF Manifests, be able to render a complete workflow, which can be executed by [runners](../../docs/runner.md). For now, we do not have a proposal, for how the rendering will be done and what kind of syntax will be used to describe the workflow.

Besides providing a syntax to define a workflow, in many cases, Content Creators would like to call other Interfaces, which are already available in OCH. For example - they are creating a workflow to provision WordPress and they need a PostgreSQL database. They have already an Interface `postgresql.install` available in OCH and they would like to use it in their Action.
Voltron must have an option, to allow Content Creators to reference another Interface in their workflow. This way, Content Creators can prepare Actions, which use the already existing platform capabilities.

Content Creator should be able to:
- reference an Interface to be called in an Action workflow,
- use and pass input parameters to Interfaces,
- reference output TypeInstances from a called Interface,
- conditionally call an Interface,
- define, which workflow artifacts are TypeInstances.

Voltron Engine must be able to:
- merge called Interfaces into the Action workflow.

For now, we want to base on the Argo workflow syntax and only extend it, to support our additional use cases. In the future, we might revisit this and change the syntax, so it is more user-friendly and can support also non-Argo runners.

### Goals

- How to reference an Interface to be called in an Action workflow.
- How to use and pass input parameters to Interfaces,
- How to reference output TypeInstances from a called Interface.
- How to conditionally call an Interface.
- How to define, which workflow artifacts are TypeInstances.
- How to merge called Interfaces into the Action workflow.

## Proposal

### How to reference an Interface to be called in an Action workflow

To reference the Interface, which has to be called, the following extensions to the Argo workflow is proposed:

- `.spec.action.args.workflow.entrypoint.templates[].steps[][].voltron-action` - defines the Interface to be called. It must be a reference to a method imported in `.spec.imports`. If this is set, the Content Creator does not have to provide a `template` field in this step.

```yaml
kind: Implementation
spec:
  imports:
    - interfaceGroupPath: cap.interface.database.postgresql
      alias: postgres
      methods:
        - name: install
          revision: 0.1.0

  action:
    runnerInterface: argo.run
    args:
      workflow:
        entrypoint: jira-install
        templates:
          - name: jira-install
            steps:
              - - name: install-db
                  voltron-action: postgres.install
```

### How to use and pass user input parameters to Interfaces

Interfaces need input parameters and Content Creators must have a way to use and also pass them to the Interfaces, they call.
We have to somehow inject the input-parameters into the workflow, so the Content Creator can reference them.

We propose, that the input parameters will be injected into the workflow as a local artifact named `input-parameters`. Fixing the artifact name, allows Voltron Engine to populate the proper artifact and also defines a standard, on how to pass input parameters to Interfaces called by the Content Creator.

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
                # This artifact, on the entrypoint template, will hold the input parameters to the Action.
                - name: input-parameters
```

During the rendering phase, the Voltron Engine will inject the `input-parameters` artifact as raw data:

```yaml
  arguments:
    artifacts:
    # Argument input-parameters added during the rendering phase.
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

Content Creators can also pass input parameters to the Interfaces, they call in their Actions.
The called Interface will be rendered into the workflow as a nested workflow. As the input parameters are passed to the workflow by the `input-parameters` artifact, Content Creators can use is to pass input parameters to the called Interface:
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
              - - name: generate-db-params
                  [...]
              - - name: install-db
                  voltron-action: postgresql.install
                  arguments:
                    artifacts:
                        # Input parameters passed to the called Interface.
                      - name: input-parameters
                        from: "{{steps.generate-db-params.outputs.params}}"
```

During the rendering this will become:

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
              - - name: generate-db-params
                  [...]
              - - name: install-db
                  template: jira-install-install-db
                  arguments:
                    artifacts:
                        # Input parameters passed to the called Interface.
                      - name: input-parameters
                        from: "{{steps.generate-db-params.outputs.params}}"
          
          # Called Interface workflow template added during the rendering phase.
          - name: jira-install-install-db
            inputs:
              artifacts:
                - name: input-parameters
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
                  voltron-action: postgresql.install
              - - name: install-jira
                  template: install-jira
                  arguments:
                    artifacts:
                      - name: "postgresql"
                        from: "{{workflow.outputs.artifacts.postgresql}}"
```

The Interface `postgresql.install` has a output TypeInstance called `postgresql` defined. To expose it to the Content Creators, we export it as a global artifact named `postgresql`.  They can then use `{{workflow.outputs.artifacts.postgresql}}` to reference the TypeInstance artifact.

### How to conditionally call an Interface

Content Creators might make their Actions self-sufficient and create the dependent TypeInstances in it, or allow the Voltron User to provide existing TypeInstances for the Action. To support this case, we need an option to conditionally call an Interface.

We decided to introduce a directive `voltron-when` to support this. Argo conditionals are evaluated during workflow execution and we need to evaluate the conditions during render-time, to do not resolve Interfaces to Implementations, and to include unnecessary workflow steps.

Only input TypeInstances can be used in the condition syntax. If the Content Creator defined an additional input TypeInstance `postgresql`, then he can make conditions based on it, for example, `voltron-when: postgresql == nil`.

For the actual implementation aspect, we propose to use the [Expr](https://github.com/antonmedv/expr) library to evaluate the condition expressions. It is used in Argo for the `depends` directive.
In the [rendering proof-of-concept](../investigation/workflow-rendering) the library [govaluate](https://github.com/Knetic/govaluate) was used, but it looks no longer maintained, based on GitHub activity.

```yaml
kind: Implementation
spec:
  additionalInput:
    typeInstances:
      # Additional input TypeInstance postgresql is defined.
      postgresql:
        typeRef:
          path: cap.type.database.postgresql.config
          revision: 0.1.0
        verbs: [ "get" ]

  action:
    runnerInterface: argo.run
    args:
      workflow:
        entrypoint: jira-install
        templates:
          - name: jira-install
            steps:
                  # Execute this step only if the postgresql TypeInstance was not provided.
              - - name: install-db
                  voltron-when: postgresql == nil
                  voltron-action: postgresql.install
```

### How to define, which workflow artifacts are TypeInstances

We need a way for the Content Creator to say, that an artifact created in the Argo workflow is a TypeInstance and is supposed to be uploaded to OCH. The workflow could use some intermediate artifacts just for handling the data flow between workflow steps. Currently, there is no way to identify the TypeInstance artifacts in the workflow.

We could enforce the Content Creator to ensure, that the TypeInstance artifact names must match with the names defined in the `.spec.additionalOutput.typeInstanceRelations`, but this would mean writing additional boilerplate steps for the Content Creator. To avoid it, we propose to define a directive `voltron-outputTypeInstances`.

The `voltron-outputTypeInstances` should be defined on workflow steps, which produce TypeInstance artifacts. Under the hood, it will create an additional workflow step, which creates a global artifact, so it can be fetched and uploaded to OCH. This also allows us to track the TypeInstances produced in a workflow.

The `voltron-outputTypeInstances` is a list of mappings between the output TypeInstance and Argo global artifacts:
```yaml
name: {output-type-instance-name}
from: {argo-global-artifact-reference}
```

```yaml
kind: Implementation
spec:
  additionalOutput:
    typeInstanceRelations:
      postgresql:
        uses:
          - helm-release

  action:
    runnerInterface: argo.run
    args:
      workflow:
        entrypoint: postgres-install
        templates:
          - name: postgres-install
            steps:
                  # This step produces Argo artifacts 'additional' and 'helm-release'.
              - - name: helm-run
                  voltron-action: cap.interface.runner.helm.run
                  voltron-output-type-instances:
                    # Artifacts mappings to the TypeInstances in .spec.additionalOutput.typeInstanceRelations
                    - name: jira-config
                      from: additional
                    - name: helm-release
                      from: helm-release
```

### How to merge called Interfaces into the Action workflow

For rendering the final workflow we use the feature of Argo workflows to execute nested workflows. As every Implementation defines an Argo workflow we can reduce the Interface calls to a nested Argo workflow.

To avoid name collisions on the workflow templates and global artifacts, we propose to prefix the template and global artifact names with a `{template.name-step.name}` prefix, based on the step, where it was called. For example, if the Interface was called in template `jira-install` and step `helm-run`, then the prefix would be `jira-install-helm-run`.

The proposed algorithm for including a nested workflow from a called Interface is described [here](../investigation/workflow-rendering/README.md).

## Example manifests with new directives

To see example manifests, with the new workflow directives check the link below. Note that these are manifests used for the PoC and the `voltron-action` contains there the full `ManifestReference`, instead of a reference to `imports` from the Implementation. This was done only to simplify the PoC.

- [PostgreSQL install](../investigation/workflow-rendering/manifests/implementations/postgres-install.yaml) - uses `voltron-outputTypeInstances` and `voltron-action`,
- [JIRA install](../investigation/workflow-rendering/manifests/implementations/jira-install.yaml) - uses `voltron-outputTypeInstances`, `voltron-when` and `voltron-action`.

## Consequences

- The workflow syntax highly depend on the Argo workflow syntax
- Add `voltron-outputTypeInstances`, `voltron-when`, `voltron-action` directives to the workflow syntax. We will need to copy-paste the Go structs, which describe Argo workflow elements and extends them.
- Argo has currently (2020.12.20) a bug, which must be fixed before the proposed rendering algorithm can work. Github ticket for this issue is [here](https://github.com/argoproj/argo/issues/4772).
