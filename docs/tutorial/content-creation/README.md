# OCF Content Creation Guide

- [OCF Content Creation Guide](#ocf-content-creation-guide)
  - [Prerequisites](#prerequisites)
  - [Types, Interfaces and Implementations](#types-interfaces-and-implementations)
  - [Define your Types and Interfaces](#define-your-types-and-interfaces)
    - [Create the Interface Group manifest](#create-the-interface-group-manifest)
    - [Create the Interface manifest](#create-the-interface-manifest)
    - [Create the Type manifests](#create-the-type-manifests)
  - [Runners](#runners)
  - [Write the Implementation for the Interface](#write-the-implementation-for-the-interface)
  - [Validate the manifests using ocftool](#validate-the-manifests-using-ocftool)
  - [Populate the manifests into OCH](#populate-the-manifests-into-och)
  - [Run your new action](#run-your-new-action)
    - [View the Action workflow in Argo UI](#view-the-action-workflow-in-argo-ui)
    - [View the Action Custom Resource](#view-the-action-custom-resource)
  - [Summary](#summary)

This guide shows the first steps on how to develop OCF content for Voltron. We will show how to:
- define new **Types** and **Interfaces**,
- create **Implementation** for the **Interfaces**,
- use other **Interfaces** in your **Implementations**,
- test the new manifests on a local development Voltron cluster

As an example, we will create OCF manifests to deploy Confluence with a PostgreSQL database.

## Prerequisites

To develop and test the created content, you will need to have a Voltron environment. To set up a local environment, install the following prerequisites:

* [Docker](https://docs.docker.com/engine/install/)
* [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [ocftool](https://github.com/Project-Voltron/go-voltron/releases/tag/v0.1.0)
* [populator](../../../cmd/populator/README.md) - For now, you need to compile it from source

Also, clone the Voltron repository with the current OCF content.
```bash
git clone https://github.com/Project-Voltron/go-voltron.git
```

Some other materials worth reading before are:
- [JIRA installation tutorial](../jira-installation/README.md) - Learn how to execute actions in Voltron.
- [Argo Workflows documentation](https://argoproj.github.io/argo-workflows/) - Voltron action syntax is based on Argo workflows, so it's highly recommended you understand what is Argo and how to create Argo workflows.
- [Voltron runners](../../runner.md) - Understand, what are Voltron runners.
- [ocftool](../../../cmd/ocftool/docs/ocftool.md) - Learn how to validate your manifests syntax.

## Types, Interfaces and Implementations

If you have some software development experience, concepts like types and interfaces should be familiar to you. In Voltron, **Types** represent different objects in the environment. These could be database or application instances, servers, but also more abstract things, like an IP address or hostname.
An actual object of a **Type** is called a **TypeInstance**.

**Interfaces** are operations, which can be executes on certain **Types** .Let's say we have a **Type** called `postgresql.config`, which represents a PostgreSQL database instance. We could have an **Interface** `postgresql.install`, which will provision a PostgreSQL instance and create a **TypeInstance** of `postgresql.config`.

**Interfaces** can be grouped into **InterfaceGroups**. **InterfaceGroups** are used to logically group the **Interfaces**. This is mostly used for presentation purposes, like to show the user all **Interfaces**, which operate on PostgreSQL instances. So if you have two **Interfaces**: `postgresql.install` and `postgresql.uninstall`, you can group them into `postgresql` InterfaceGroup.

Of course, there are multiple ways to create a PostgreSQL instance: you can create it on a public cloud or on-premise, and you could deploy it as a virtual machine or as a Kubernetes StatefulSet. To cover these scenarios, Voltron allows defining multiple **Implementations** of some **Interfaces**. For example:
- `aws.postgresql.install` **Implementation** of the `postgresql.install` **Interface** will deploy AWS RDS instances, whereas
- `bitnami.postgresql.install` **Implementation** will deploys a PostgreSQL Helm chart on Kubernetes.

## Define your Types and Interfaces

Let's try to create manifests required to define a capability to install Confluence servers. We will need to create the following entities:
- `confluence.config` **Type** - Represents a Confluence server.
- `confluence.install-input` **Type** - Represents input parameters needed to install a Confluence server.
- `confluence` **InterfaceGroup** - Groups **Interfaces** from the `confluence` group, e.g. if you have `confluence.install` and `confluence.uninstall` **Interfaces**.
- `confluence.install` **Interface** - An operation, which installs Confluence servers. You can think of it as a function:
```
confluence.install(confluence.install-input) -> confluence.config
```

### Create the Interface Group manifest

First, we need to create an **InterfaceGroup** manifest, which groups **Interfaces** coresponding to some application.
Let's create a InterfaceGroup called `cap.interface.productivity.confluence`, which will group **Interfaces** operating on Confluence instances.

In `och-content/interface/productivity/`, create a file called `confluence.yaml`, with the following content:

<details>
  <summary>och-content/interface/productivity/confluence.yaml</summary>

```yaml
ocfVersion: 0.0.1
revision: 0.1.0
kind: InterfaceGroup
metadata:
  prefix: cap.interface.productivity
  name: Confluence
  displayName: "Confluence Server"
  description: "Confluence is a document collaboration tool"
  documentationURL: https://support.atlassian.com/bitbucket-cloud/
  supportURL: https://support.atlassian.com/bitbucket-cloud/
  iconURL: https://wac-cdn.atlassian.com/dam/jcr:d6e2d2db-e58a-40f7-9d1a-d6d22a335c96/Confluence-blue.svg
  maintainers:
    - email: your.email@example.com
      name: your-name
      url: your-website

signature:
  och: eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9
```
</details>

> The `signature` field is required, but currently we haven't yet implemented a signing mechanism, so you can put a dummy value there.

### Create the Interface manifest

After we have the **InterfaceGroup**, let's create the **Interface** for installing Confluence.
Create the directory `och-content/interface/productivity/confluence`.

Inside this directory, create a file `install.yaml` with the following content:

<details>
  <summary>och-content/interface/productivity/confluence/install.yaml</summary>

```yaml
ocfVersion: 0.0.1
revision: 0.1.0
kind: Interface
metadata:
  prefix: cap.interface.productivity.confluence
  name: install
  displayName: "Install Confluence"
  description: "Confluence is a document collaboration tool"
  documentationURL: https://support.atlassian.com/confluence/
  supportURL: https://support.atlassian.com/confluence/
  iconURL: https://wac-cdn.atlassian.com/dam/jcr:d6e2d2db-e58a-40f7-9d1a-d6d22a335c96/Confluence-blue.svg
  maintainers:
    - email: your.email@example.com
      name: your-name
      url: your-website

spec:
  input:
    parameters:
      input-parameters:   # the Interface requires an input-parameter of Type "cap.type.productivity.confluence.install-input"
        jsonSchema:
          value: |-
            {
              "$schema": "http://json-schema.org/draft-07/schema",
              "$ocfRefs": {
                "inputType": {
                  "name": "cap.type.productivity.confluence.install-input",
                  "revision": "0.1.0"
                }
              },
              "allOf": [ { "$ref": "#/$ocfRefs/inputType" } ]
            }
  output:
    typeInstances:
      confluence-config:    # the Interface outputs TypeInstances of Type "cap.type.productivity.confluence.config"
        typeRef:
          path: cap.type.productivity.confluence.config
          revision: 0.1.0

signature:
  och: eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9
```
</details>

The `spec.input` property defines inputs, required by the **Interface**. There are two types of inputs:
- `spec.input.parameters` - User provided input parameters, i.e. these could be configuration parameters required by the operation,
- `spec.input.typeInstances` - input **TypeInstances**, i.e. a PostgreSQL database, which is needed for an application.

The `spec.output` property defines the **TypeInstance**, which this **Interface** returns.

Although Confluence needs a database, we don't specify it as an input argument here. That is because, we leave selecting a database to the **Implementation**.

### Create the Type manifests

Now we need to define the two **Types**, which we use in our **Interface**: `cap.type.productivity.confluence.install-input` and `cap.type.productivity.confluence.config`.

<details>
  <summary>och-content/type/productivity/confluence/install-input.yaml</summary>

```yaml
ocfVersion: 0.0.1
revision: 0.1.0
kind: Type
metadata:
  name: install-input
  prefix: cap.type.productivity.confluence
  displayName: Confluence installation input
  description: Defines installation parameters for Confluence
  documentationURL: https://support.atlassian.com/confluence-cloud/
  supportURL: https://www.atlassian.com/software/confluence
  iconURL: https://wac-cdn.atlassian.com/dam/jcr:d6e2d2db-e58a-40f7-9d1a-d6d22a335c96/Confluence-blue.svg
  maintainers:
    - email: your.email@example.com
      name: your-name
      url: your-website

spec:
  jsonSchema:
    value: |-
      {
        "$schema": "http://json-schema.org/draft-07/schema",
        "type": "object",
        "title": "The schema for Jira configuration",
        "required": [
            "host"
        ],
        "$ocfRefs": {
          "hostname": {
            "name": "cap.core.type.networking.hostname",
            "revision": "0.1.0"
          }
        },
        "properties": {
          "host": {
            "$ref": "#/$ocfRefs/hostname"
          }
        },
        "additionalProperties": true
      }

signature:
  och: eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9
```
</details>

<details>
  <summary>och-content/type/productivity/confluence/config.yaml</summary>

```yaml
ocfVersion: 0.0.1
revision: 0.1.0
kind: Type
metadata:
  name: config
  prefix: cap.type.productivity.confluence
  displayName: Confluence instance config
  description: Defines configuration for Confluence instance
  documentationURL: https://support.atlassian.com/confluence-cloud/
  supportURL: https://www.atlassian.com/software/confluence
  iconURL: https://wac-cdn.atlassian.com/dam/jcr:d6e2d2db-e58a-40f7-9d1a-d6d22a335c96/Confluence-blue.svg
  maintainers:
    - email: your.email@example.com
      name: your-name
      url: your-website

spec:
  jsonSchema:
    value: |-
      {
        "$schema": "http://json-schema.org/draft-07/schema",
        "type": "object",
        "title": "The schema for Jira configuration",
        "required": [
            "version"
        ],
        "$ocfRefs": {
          "semVer": {
            "name": "cap.core.type.versioning.semver",
            "revision": "0.1.0"
          }
          "hostname": {
            "name": "cap.core.type.networking.hostname",
            "revision": "0.1.0"
          }
        },
        "properties": {
          "version": {
            "$ref": "#/$ocfRefs/semVer"
          }
          "host": {
            "$ref": "#/$ocfRefs/hostname"
          }
        },
        "additionalProperties": true
      }

signature:
  och: eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9
```
</details>

The **Type** values are described using [JSON Schema](https://json-schema.org/).

> Currently the **Type** manifests are not used in Voltron to validate the data of the inputs and outputs. Validation of the data will be added later on, although
> it is still useful to define the **Types** to document the schema of the data.

## Runners

The Action execution is handled by runners. For now we provide the following runners:
- [Argo Workflow Runner](../../../cmd/argo-runner/README.md)
- [Helm Runner](../../../cmd/helm-runner/README.md)
- [CloudSQL Runner](../../../cmd/cloudsql-runner/README.md)
- [Terraform Runner](../../../cmd/terraform-runner/README.md)

To check the schema of the runner input, you have to look in the [`och-content/type/runner`](../../../och-content/type/runner) directory. You can find there the JSON schema and an example input for the runner.

You can read more about runners in [this document](../../runner.md).

## Write the Implementation for the Interface

> The syntax used to describe the workflows in **Implementations** is based on [Argo](https://argoproj.github.io/argo/).
> It's highly recommended you read their documentation and understand what is Argo and how to create Argo workflows, before writing OCF Implementations.

After we defined the **Interfaces** and the **Types**, we can write an **Implementation** of `confluence.install`. Our **Implementation** will use a PostgreSQL database, which will be provided by another **Interface**, which is already available in Voltron. We also allow users to provide his own PostgreSQL instance **TypeInstance**.

Create a file `och-content/implementation/atlassian/confluence/install.yaml` with the following content:

<details>
  <summary>och-content/implementation/atlassian/confluence/install.yaml</summary>

```yaml
ocfVersion: 0.0.1
revision: 0.1.0
kind: Implementation
metadata:
  prefix: cap.implementation.atlassian.confluence
  name: install
  displayName: Install Confluence
  description: Action which installs Confluence via Helm chart
  documentationURL: https://github.com/javimox/helm-charts/tree/master/charts/confluence-server
  supportURL: https://mox.sh/helm/
  license:
    name: "Apache 2.0"
  maintainers:
    - email: your.email@example.com
      name: your-name
      url: your-website

spec:
  appVersion: "2.x.x"

  # Here we define an additional, optional input TypeInstance of Type "cap.type.database.postgresql.config".
  additionalInput:
    typeInstances:
      postgresql:
        typeRef:
          path: cap.type.database.postgresql.config
          revision: 0.1.0
        verbs: [ "get" ]

  # We can also define additional output TypeInstances, which our Implementation creates.
  # You can use in case your Implementation creates some additional resources,
  # which are not mentioned in the Interface, but have to be managed by Voltron.
  #additionalOutput:
  #  typeInstances:
  #    some-additional-resource:
  #      typeRef:
  #        path: cap.type.instance
  #        revision: 0.1.0
  
  # We specify which TypeInstances are uploaded to OCH as TypeInstances, along with theirs relations.
  # We can specify here both required and optional output TypeInstances for a given manifest.
  # You can also specify TypeInstances, which are defined in voltron-outputTypeInstances in steps.
  # If a given Argo artifact is not mentioned here, it won't be created in OCH as TypeInstance.
  outputTypeInstanceRelations:
      confluence-config:
        uses:
          - confluence-helm-release
          - postgresql
          - database
  # If a given TypeInstance should be uploaded to OCH, but it doesn't have any dependencies,
  # you can specify it such as:
  #   some-additional-resource: {}

  # Here we say, we implement the Interface, we defined before.
  implements:
    - path: cap.interface.productivity.confluence.install
      revision: 0.1.0

  requires:
    cap.core.type.platform:
      oneOf:
        - name: kubernetes
          revision: 0.1.0

  # In this section we need to import all Interfaces, we will use in our Implementation.
  # These can be Voltron runner or other actions.
  imports:
    - interfaceGroupPath: cap.interface.runner.argo
      alias: argo
      methods:
        - name: run
          revision: 0.1.0
    - interfaceGroupPath: cap.interface.runner.helm
      alias: helm
      methods:
        - name: run
          revision: 0.1.0
    - interfaceGroupPath: cap.interface.database.postgresql
      alias: postgresql
      methods:
        - name: install
          revision: 0.1.0
        - name: create-db
          revision: 0.1.0
    - interfaceGroupPath: cap.interface.jinja2
      alias: jinja2
      methods:
        - name: template
          revision: 0.1.0

  action:
    runnerInterface: argo.run
    args:
      workflow:
        entrypoint: main
        templates:
          - name: main
            inputs:
              artifacts:
                # Voltron Engine injects the 'input-parameters' artifacts into the workflow entrypoint.
                # It contains the Interface parameters, in our case it is `confluence.install-input`.
                - name: input-parameters
                # You need to specify all optional TypeInstances that can be consumed by the workflow.
                # Later, the user or other workflow can inject own TypeInstance which fulfills schema for this TypeInstance.
                # You as a Workflow developer, can refer to them in the `voltron-when` statements.
                - name: postgresql # Same name as defined under the `additionalInput.typeInstances` property.
                  optional: true
            # Under `outputs`, you need to specify all outputs defined under Interface that this workflow implements.
            # As a result, others can refer to the Action output via "{{ steps.<step_name>.outputs.artifacts.<artifact_name> }}".
            # Same as we do with the postgres step in this workflow.
            outputs:
                - name: confluence-config # Same name as defined in Interface
                  from: "{{steps.helm-run.outputs.artifacts.additional}}" # Instructs Argo which artifact holds the Confluence config
            steps:
              # If the postgresql input argument was not provided, then create it
              # using the imported 'postgresql.install' Interface.
              # Otherwise, it is replaced with the "mock" step emitting the input argument which satisfied this step as a step output.
              - - name: install-db
                  voltron-action: postgresql.install
                  voltron-when: postgresql == nil
                  voltron-outputTypeInstances:
                    - name: postgresql                # We have to define the output TypeInstance 'postgresql' from this step
                      from: postgresql
                  arguments:
                    artifacts:
                      - name: input-parameters
                        raw:
                          data: |
                            superuser:
                              username: superuser
                              password: okon
                            defaultDBName: postgres

              # Create the database for Confluence in our PostgreSQL instance
              # using the imported 'postgresql.create-db' Interface
              - - name: create-db
                  voltron-action: postgresql.create-db
                  voltron-outputTypeInstances:
                    - name: database                  # Defining the output TypeInstance 'database'
                      from: database
                  arguments:
                    artifacts:
                      - name: postgresql
                        from: "{{steps.install-db.outputs.artifacts.postgresql}}" # Refer to output from `install-db` step. 
                                                                                  # Accessing the `install-db` output is possible even if the step was 
                                                                                  # satisfied by the input argument and replaced with the "mock" step.
                      - name: database-input
                        raw:
                          data: |
                            name: confluencedb 
                            owner: superuser

              # Here we prepare the input for the Helm runner. In the next two steps,
              # we use Jinja2 to render the input and fill the required parameters.
              # In the future there might be a better way to do this.
              - - name: render-helm-args
                  voltron-action: jinja2.template
                  arguments:
                    artifacts:
                      - name: template
                        raw:
                          data: |
                            command: "install"
                            generateName: true
                            chart:
                              name: "confluence-server"
                              repo: "https://helm.mox.sh"
                            output:
                              goTemplate:
                                version: "{{ .Values.image.tag }}"
                                host: "{{ template \"confluence-server.fullname\" . }}"
                            values:
                              postgresql:
                                enabled: false
                              databaseConnection:
                                host: "<@ host @>"
                                user: "<@ superuser.username @>"
                                password: "<@ superuser.password @>"
                                database: "<@ name @>"
                              ingress:
                                enabled: true
                                hosts:
                                - host: confluence.voltron.local
                                  paths: ['/']
                      - name: input-parameters
                        from: "{{steps.install-db.outputs.artifacts.postgresql}}"

              - - name: fill-params-in-helm-args
                  voltron-action: jinja2.template
                  arguments:
                    artifacts:
                      - name: template
                        from: "{{steps.render-helm-args.outputs.artifacts.render}}"
                      - name: input-parameters
                        from: "{{steps.create-db.outputs.artifacts.database}}"

              # Execute the Helm runner, with the input parameters created in the previous step.
              # This will create the Helm chart and deploy our Confluence instance
              - - name: helm-run
                  voltron-action: helm.run
                  voltron-outputTypeInstances:
                    - name: confluence-config         # Defining the output TypeInstance 'confluence-config'
                      from: additional
                    - name: confluence-helm-release   # Defining the output TypeInstance 'confluence-helm-release'
                      from: helm-release
                  arguments:
                    artifacts:
                      - name: input-parameters
                        from: "{{steps.fill-params-in-helm-args.outputs.artifacts.render}}"
                      - name: runner-context
                        from: "{{workflow.outputs.artifacts.runner-context}}"

signature:
  och: eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9
```
</details>

Let's take a look on the **Implementation** YAML. **Implementation** has the following properties in the `spec` property:

| Property                      | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `appVersion`                  | Application versions, which this **Implementation** supports.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `additionalInput`             | Additional input for the **Implementation**, compared to the **Interface**. In our case, here we define the `postgresql.config`, as our **Implementation** uses a PostgreSQL instance for Confluence.                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `additionalOutput`            | This section defines any additional **TypeInstances**, which are created in this **Implementation**, compared to the **Interface**. In our **Implementation**, we create a database in the database instance with the `postgresql.create-db` **Interface**, which outputs an `postgresql.database` **TypeInstance**. We have to write this down in `additionalOutput`, so Voltron will resolve this **TypeInstance** metadata for uploading it to OCH.                                                                                                                                                                                      |
| `outputTypeInstanceRelations` | Specifies all output TypeInstances to upload to OCH with theirs relationships between them. Only the TypeInstances created in this Implementation have to be mentioned here. If a TypeInstances in created in another action and brought into the context with `voltron-outputTypeInstances`, then it should not be defined here.                                                                                                                                                                                                                                                                                                                                                                                               |
| `implements`                  | Defines which **Interfaces** are implemented by this **Implementation**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `requires`                    | List of system prerequisites that need to be present in the environment managed by Voltron to use this **Implementation**. In our example, we will deploy Confluence as a Helm chart on Kubernetes, which means we need a Kubernetes cluster. Requirement items can specify `alias` and be used inside workflow under `{{workflow.outputs.artifacts.{alias}}}`, where `{alias-name}` is the alias. A TypeInstance with alias is injected into the workflow based on Policy configuration. To learn more, see the [TypeInstance Injection](../../policy-configuration.md#typeinstance-injection) paragraph in Policy Configuration document. |
| `imports`                     | Here we define all other **Interfaces**, we use in our **Implementation**. We can then refer to them as `'<alias>.<method-name>'`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `action`                      | Holds information about the actions that is executed. In the case of the Argo workflow Runner, in this section we define the Argo workflow, which is executed in this **Implementation**.                                                                                                                                                                                                                                                                                                                                                                                                                                                   |

> You can notice, that `confluence-config` (which is the `additional` output TypeInstance from `helm.run`) is defined in the `outputTypeInstanceRelations`, although it was created in `helm.run`. The `additional` from `helm.run` is specially, because `helm.run` does not know the Type of TypeInstances, so it's not defined in `helm.run` Implementation, but must be defined in the caller Implementation. In the future, we will improve the syntax, so it will be more clear, which TypeInstances need a separate entry in `outputTypeInstanceRelations` and which don't.

The workflow syntax is based on [Argo](`https://argoproj.github.io/argo/`), with a few extensions introduced by Voltron. These extensions are:

| Property                                          | Description                                                                                                                                                                                                                                                                                                           |
| ------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `.templates.steps[][].voltron-when`               | Allows for conditional execution of a step, based on an expression with an input workflow artifacts arguments. You can make assertions on artifacts defined under `inputs.arguments.artifacts` for a given template. It supports the syntax defined here: [antonmedv/expr](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md). |
| `.templates.steps[][].voltron-action`             | Allows to import another **Interface**. In our example, we use this to provision PostgreSQL with `postgresql.install` **Interface**.                                                                                                                                                                                  |
| `.templates.steps[][].voltron-outputTypeInstance` | A list of **TypeInstances**, from the called action, which are brought into the context of this **Implementations**. The `from` property must match the name of the output from the called Action. You can then use it in the Implementations `outputTypeInstanceRelations`, when defining relations between TypeInstances. |

Let's go through the **Implementation** and try to understand, what is happening in each step of the action. Our Confluence installation uses a PostgreSQL database. We defined an additional input `postgresql` of type `cap.type.database.postgresql.config`. Additional inputs are optional, so we need to handle the scenario, where no **TypeInstance** for `postgresql`  was provided. The first workflow step `install-db` is conditionally using the `postgresql.install` **Interface** to create an PostgreSQL instance.

> The `input-parameters` for `postgresql.install` are hardcoded in this example. In a real workflow, they should be generated or taken from the `input-parameters` for this **Implementation**.

In the next step we are creating a database for the Confluence server. If you look at the **Interface** definition of [`cap.interface.database.postgresql.create-db`](och-content/interface/database/postgresql/create-db.yaml), you will see, that it requires a `postgresql` **TypeInstance** of **Type** [`cap.type.database.postgresql.config`](och-content/type/database/postgresql/config.yaml) and input parameters [`cap.type.database.postgresql.database-input`](och-content/type/database/postgresql/database-input.yaml), and outputs a `database` **TypeInstance** of **Type** [`cap.type.database.postgresql.database`](och-content/type/database/postgresql/database.yaml). In the step, we are providing the inputs to the **Interface** via the `.arguments.artifacts` field. We also have to map the output of this step to our output definitions in `additionalOutput` and the implemented **Interface** in the `voltron-outputTypeInstances` field.

The `render-helm-args` and `fill-params-in-helm-args` steps are used to prepare the input parameters for the `helm.run` **Interface**. Jinja template engine is used here to render the Helm runner arguments with the required data from the `postgresql` and `database` **TypeInstances**. Those steps don't create any **TypeInstances** and serve only the purpose of creating the input parameters for the Helm runner.
You can check the schema of the Helm runner args in the [Type manifest](../../../och-content/type/runner/helm/run-input.yaml).

> To create the input parameters for `helm.run` we have to use data from two artifacts. As the current `jinja.run` **Interface** consumes only a template and a single variables input, we have to perform this operation twice. To seperate the variables substituted in the first and second operation, we escape some parts using `{% raw %} ... {% endraw %}`, which is removed in the first templating operation and will be processed in the second operation.
>
> In the future we might improve the ways, on how to process artifacts in the workflow.

The last step launches the Helm runner, deploys the Confluence server and creates the `confluence-config` and `confluence-helm-release` **TypeInstances**. The `confluence-config` **TypeInstance** data was provided by the Helm runner in the `additional` output artifacts from this step. Check the Helm runner documentation, on how the `additional` output is created.

Note the `runner-context` argument, which provided the context for the runner. Voltron Engine injects a global artifact `workflow.outputs.artifacts.runner-context` into the workflow, so if you use a runner, which needs the runner context, you can to do it using:
```yaml
arguments:
  artifacts:
    - name: runner-context
      from: "{{workflow.outputs.artifacts.runner-context}
```
To verify, if a runner needs the context, check the **Interface** of the runner (eg. [Interface for Helm runner](../../../och-content/interface/runner/helm/run.yaml)).

## Validate the manifests using ocftool

You can use the `ocftool` to validate the manifests you created. The `ocftool validate` command checks the manifests against JSON schemas and can tell you, if your manifests are correct.

> For now the `ocftool` does not verify the content of the `action` property in **Implementations**. It will not verify, that your workflow is correct and will execute properly.

To verify all your manifests in `och-content` directory, execute:
```
ocftool validate och-content/**/*.yaml
```

You can read more about the `ocftool` [here](../../../cmd/ocftool/README.md).

## Populate the manifests into OCH

After we have the manifests ready, we can start our local Voltron environment. In the root of the cloned `go-voltron` repository run:
```
ENABLE_POPULATOR=false make dev-cluster
```

This can take a few minutes. We disabled the populator sidecar in OCH public, as we will populate the data from our local repository using the populator.

> You can read more about the populator, how to compile and use it, in this [README](../../../cmd/populator/README.md).

To populate the data, you will need to first set up port-forwarding to the Neo4j database service:
```
kubectl port-forward -n neo4j svc/neo4j-neo4j 7474 7687
```

Then populate the data, with the populator:
```
APP_JSONPUBLISHADDR=<your-local-docker-ip-address> APP_MANIFESTS_PATH=och-content ./populator .

APP_JSONPUBLISHADDR=http://172.17.0.1 APP_MANIFESTS_PATH=och-content populator .
```

## Run your new action

Now we will create the action, to trigger the Confluence installation. Open https://gateway.voltron.local/ in your browser.
Then copy the following queries, variables and HTTP headers to the GraphQL playground:

<details>
  <summary>GraphQL queries</summary>

```graphql
mutation CreateAction($in: ActionDetailsInput!) {
  createAction(in: $in) {
    name
    status {
      phase
      message
    }
  }
}

query GetAction($actionName: String!) {
  action(name: $actionName) {
    name
    status {
      phase
      message
    }
  }
}

mutation RunAction($actionName: String!) {
  runAction(name: $actionName) {
    name
    status {
      phase
      message
    }
  }
}

mutation DeleteAction($actionName: String!) {
  deleteAction(name: $actionName) {
    name
  }
}
```
</details>

<details>
  <summary>Query variables</summary>

```json
{
  "actionName": "install-confluence",
  "in": {
    "name": "install-confluence",
    "actionRef": {
      "path": "cap.interface.productivity.confluence.install",
      "revision": "0.1.0"
    }
  }
}
```
</details>

<details>
  <summary>HTTP headers</summary>

```json
{
  "Authorization": "Basic Z3JhcGhxbDp0MHBfczNjcjN0"
}
```
</details>

![graphql-playground](./assets/graphql-playground.png)

Execute the `CreateAction` mutation. This will create the Action resource in Voltron. You can use the `GetAction` query to check the status of the Action. Wait till it is in the `READY_TO_RUN` phase.

After it is in the `READY_TO_RUN` phase, you can see the workflow, which will be executed in the `renderedAction` field. To run the Action, execute the `RunAction` mutation. Use the `GetAction` query to monitor the status of the Action.

### View the Action workflow in Argo UI

You can also check the status of the Action by monitoring the workflow for the Action on the Argo UI. This can give you useful information, when debugging your Action.
To get access to the Argo UI, execute the following command to set up port-forwarding to Argo:
```
kubectl port-forward -n argo svc/argo-server 2746
```

Now you can access the Argo UI with your browser by opening http://127.0.0.1:2746.

![argo-ui](./assets/argo-ui.png)

### View the Action Custom Resource

You can get useful information about your action using `kubectl`. You can check the `actions.core.projectvoltron.dev` custom resource to get information about your action:
```bash
$ kubectl describe actions.core.projectvoltron.dev install-confluence
Name:         install-confluence
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  core.projectvoltron.dev/v1alpha1
[...]
Status:
  Last Transition Time:  2021-02-08T17:17:27Z
  Message:               Rendering runner action
  Observed Generation:   1
  Phase:                 BeingRendered
Events:
  Type     Reason                Age                  From               Message
  ----     ------                ----                 ----               -------
  Normal   BeingRendered         3m2s                 action-controller  Rendering runner action
  Warning  Render runner action  2s (x15 over 2m58s)  action-controller  while resolving Implementation for Action: while rendering Action: No implementation found for "cap.interface.productivity.confluence.install"
```

In the case above, we can see that the action rendering is failing, because the Voltron Engine is not able to find the `cap.interface.productivity.confluence.install` **Implementation** in OCH.

## Summary

In this guide we went through different OCF manifests and their syntax. We created manifests which added a capability to install Confluence server instances. We also showed, how you can test the manifests you are creating and where to check for useful information, when debugging your action.
