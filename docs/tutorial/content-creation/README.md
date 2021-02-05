# OCF Content Creation Guide

- [OCF Content Creation Guide](#ocf-content-creation-guide)
  - [Prerequisites](#prerequisites)
  - [Types, Interfaces and Implementations](#types-interfaces-and-implementations)
  - [Define your Types and Interfaces](#define-your-types-and-interfaces)
  - [Write the Implementation for the Interface](#write-the-implementation-for-the-interface)
  - [Populate the manifests into OCH](#populate-the-manifests-into-och)
  - [Run your new action](#run-your-new-action)

This guide shows first steps on how to develop OCF content for Voltron. We will show how to:
- define new Types and Interfaces,
- create Implementation for the Interfaces,
- use other Interfaces in your Implementations,
- test the new manifests on a local development Voltron cluster

As an example, we will create OCF manifests to deploy Confluence with a PostgreSQL database.

## Prerequisites

To develop and test the created content, you will need to have a Voltron environment. To set up a local environment, install the following prerequisites:

* [Docker](https://docs.docker.com/engine/install/)
* [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [ocftool](https://github.com/Project-Voltron/go-voltron/releases/tag/v0.1.0)
* [populator](TBD) TODO: release or give instruction on how to compile it from source


Also, clone the Voltron repository with the current OCF content.
```bash
git clone https://github.com/Project-Voltron/go-voltron.git
```

## Types, Interfaces and Implementations

If you have some software development experience, concepts like types and interfaces should be familliar to you. In Voltron, Types represent different objects in the environment. These could be database or application instances, servers, but also more abstract things, like an IP address or hostname.
An actual object of a Type is called a TypeInstance.

> TODO: tutaj może obrazek z VMką na GCP i pokazanie czym jest type instancja a czym typ

Operations, which can be executed on some Types are defined by Interfaces. Let's say we have an Type called `postgresql.config`, which represents a PostgreSQL database instance. We could have an Interface `postgresql.install`, which will provision PostgreSQL instances and create TypeInstances of `postgresql.config`.

> TODO: tutaj obrazek z tym co robią interfejsy

Of course, there could be multiple ways, how to create an PostgreSQL instance. You could create it on some public cloud or on-premise. You could deploy it as a virtual machine or as a Kubernetes StatefulSet. To cover these scenarios, Voltron allows to define multiple Implementations of some Interfaces. So we could have a `aws.postgresql.install` Implementation of the `postgresql.install` Interface, which deploys AWS RDS instances or `bitnami.postgresql.install`, which deploys a PostgreSQL Helm chart on Kubernetes.

> TODO: obrazek z roznymi implementacjami dla postgresql.install

## Define your Types and Interfaces

Let's try to create manifests required to define an capability to install Confluence servers. We will need to create the following entities:
- `confluence.config` Type - Represents a Confluence server.
- `confluence.install-input` Type - Represents input parameters needed to install a Confluence server.
- `confluence.install` Interface - An operation, which installs Confluence servers. You can think of it as a function:
```
confluence.install(confluence.install-input) -> confluence.config
```
- `confluence` InterfaceGroup - Groups Interfaces from the `confluence` group, e.g. if you have `confluence.install` and `confluence.uninstall` Interfaces.

As first, you need to create an **InterfaceGroup** manifest, which groups Interfaces coresponding to some application.
Let's create a InterfaceGroup called `cap.interface.productivity.confluence`, which will group Interfaces operating on Confluence instances. In `och-content/interface/productivity/`, create a file called `confluence.yaml`, with the following content:
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
  iconURL: https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png
  maintainers:
    - email: your.email@example.com
      name: your-name
      url: your-website

signature:
  och: eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9
```

> The `signature` field is required, but currently we don't have implemented yet a signing mechanism. You can put a dummy value there.

After we have the InterfaceGroup, let's create the Interface, for installing Confluence.
Create the directory `./och-content/interface/productivity/confluence`. Inside this directory, create a file `install.yaml` with the following content:
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
  iconURL: https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png # TODO change this
  maintainers:
    - email: team-dev@projectvoltron.dev
      name: Voltron Dev Team
      url: https://projectvoltron.dev

spec:
  input:
    parameters:
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
      jira-config:
        typeRef:
          path: cap.type.productivity.confluence.config
          revision: 0.1.0

signature:
  och: eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9
```

The `spec.input` key defines inputs, required by the Interfaces. There are two types of inputs:
- `spec.input.parameters` - User provided input parameters, i.e. these could be configuration parameters required by the operation,
- `spec.input.typeInstances` - input TypeInstances, i.e. a PostgreSQL database, which is needed for an application.

Although Confluence needs an database, we don't specify is as an input argument here. That is because, we leave selecting a database to the Implementation.

Now we need to define the two Types, which we use in our Interface: `cap.type.productivity.confluence.install-input` and `cap.type.productivity.confluence.config`.

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
  iconURL: https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png
  maintainers:
    - email: team-dev@projectvoltron.dev
      name: Voltron Dev Team
      url: https://projectvoltron.dev

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
  iconURL: https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png
  maintainers:
    - email: team-dev@projectvoltron.dev
      name: Voltron Dev Team
      url: https://projectvoltron.dev

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

The Types are described using [JSON Schema](https://json-schema.org/).

> Currently the Type manifests are not used in Voltron to validate the data of the inputs and outputs. Validation of the data will be added later on, although
> it is still useful to define the Types to document the schema of the data.

## Write the Implementation for the Interface

After we defined the Interfaces and the Types, we can write a Implementation of `confluence.install`. Our Implementation will use a PostgreSQL database, which will be provided by an another Interface, which is already available in Voltron. We will also allow the user to provide his own PostgreSQL instance TypeInstance. Create a file `och-content/implementation/atlassian/confluence/install.yaml` with the following content:
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
    - email: team-dev@projectvoltron.dev
      name: Voltron Dev Team
      url: https://projectvoltron.dev

spec:
  appVersion: "2.x.x"

  additionalInput:
    typeInstances:
      postgresql:
        typeRef:
          path: cap.type.database.postgresql.config
          revision: 0.1.0
        verbs: [ "get" ]

  additionalOutpu t:
    typeInstances:
      confluence-helm-release:
        typeRef:
          path: cap.type.helm.chart.release
          revision: 0.1.0
      database:
        typeRef:
          path: cap.type.postgresql.database
          revision: 0.1.0
    typeInstanceRelations:
      confluence-config:
        uses:
          - confluence-helm-release
          - postgresql
          - database

  implements:
    - path: cap.interface.productivity.confluence.install
      revision: 0.1.0

  requires:
    cap.core.type.platform:
      oneOf:
        - name: kubernetes
          revision: 0.1.0

  imports:
    - interfaceGroupPath: cap.interface.runner.argo
      alias: argo
      methods:
        - name: run
          revision: 0.1.0
    - interfaceGroupPath: cap.interface.runner.helm
      alias: helm
      appVersion: 3.x.x
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
            # Voltron Engine will inject the 'input-parameters' artifacts into the workflow entrypoint.
            # It contains the Interface parameters, in our case it is `confluence.install-input`.
            inputs:
              artifacts:
                - name: input-parameters
            steps:
              # If the postgresql TypeInstance was not provided, then create it
              # using the imported 'postgresql.install' Interface.
              - - name: install-db
                  voltron-action: postgresql.install
                  voltron-when: postgresql == nil
                  voltron-outputTypeInstances:
                    - name: postgresql                # Defining the output TypeInstance 'postgresql'
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
                        from: "{{workflow.outputs.artifacts.postgresql}}"
                      - name: database-input
                        raw:
                          data: |
                            name: confluencedb 
                            owner: superuser

              # Here we prepare the input for the Helm runner. In the next two steps,
              # we use Jinja2 to render the input and fill the required parameters.
              # In the future there might be better way to do this.
              - - name: render-helm-args
                  voltron-action: jinja2.template
                  arguments:
                    artifacts:
                      - name: template
                        raw:
                          data: |
                            context:
                              name: "confluence-helm-release"
                              dryRun: false
                              timeout: "10m"
                              platform:
                                namespace: "default"
                            args:
                              command: "install"
                              generateName: true
                              chart:
                                name: "confluence-server"
                                repo: "https://helm.mox.sh"
                              output:{% raw %}
                                goTemplate:
                                  version: {{ '"{{ .Values.image.tag }}"' }}
                                  host: {{ "'{{ template \"confluence-server.fullname\" . }}'" }}{% endraw %}
                              values:
                                postgresql:
                                  enabled: false
                                databaseConnection:
                                  host: "{{ host }}"
                                  user: "{{ superuser.username }}"
                                  password: "{{ superuser.password }}"
                                  {% raw %}database: "{{ name }}"{% endraw %}
                                ingress:
                                  enabled: true
                                  hosts:
                                  - host: confluence.voltron.local
                                    paths: ['/']
                      - name: input-parameters
                        from: "{{workflow.outputs.artifacts.postgresql}}"

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

signature:
  och: eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9
```

Let's take a look on the Implementation YAML. Implementation has the following fields in the `spec` field:
- `appVersion` - Application versions, which this Implementation supports.
- `additionalInput` - Additional input for the Implementation, compared to the Interface. In our case, here we define the `postgresql.config`, as our Implementation uses a PostgreSQL instance for Confluence.
- `additionalOutput` - This section defines any additional TypeInstances, which are created in this Implementation, compared to the Interface. In example, in our Implementation, we create an database in the database instance with the `postgresql.create-db` Interface, which outputs an `postgresql.database` TypeInstance. We have to write this down in `additionalOutput`, so Voltron will upload this TypeInstance to OCH and save the dependency graph.
- `implements` - Defines, which Interfaces are implemented by this Implementation.
- `requires` - Defines additional constraints, which must be met by the Voltron environment, so this Implementation can be used. In our example, we will deploy Confluence as a Helm chart on Kubernetes, which means, we need an Kubernetes cluster.
- `imports` - Here we define all other Interfaces, we use in our Implementation. We can then refer to them as `'<alias>.<method-name>'`.
- `action` - In this section we define the workflow, which is executed in this Implementation.

The workflow syntax is based on [Argo](`https://argoproj.github.io/argo/`), with a few extensions introduced by Voltron. These extensions are:
- `.templates.steps[][].voltron-when` - Allows for conditional execution of a step. You can make assertions on the input TypeInstances available in the Implementation. It supports the syntax defined here: [antonmedv/expr](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md).
- `.templates.steps[][].voltron-action` - Allows to import an another Interface. In our example, we use this to provision on PostgreSQL with `postgresql.install` Interface, if it not was provided as a TypeInstance or deploy Helm charts using `helm.run` Interface.
- `.templates.steps[][].voltron-outputTypeInstance` - A list of TypeInstances, from the Implementations outputs, which are created in this step. The `name` must match with the output name defined in the implemented Interface or Implementations `additionalOutput` and `from` is the name of the Argo output artifacts from this step.

Our Confluence installation uses a PostgreSQL database. We defined an additional input `postgresql` of type `cap.type.database.postgresql.config`. Additional inputs are optional, so we need to handle the scenario, where no TypeInstance for `postgresql`  was provided. The first workflow step `install-db` is conditionally using the `postgresql.install` Interface to create an PostgreSQL instance.

> The `input-parameters` for `postgresql.install` are hardcoded in this example. In a real workflow, they should be generated or taken from the `input-parameters` for this Implementation.

In the next step we are creating an database for the Confluence server. If you look at the Interface definition of [`cap.interface.database.postgresql.create-db`](och-content/interface/database/postgresql/create-db.yaml), you will see, that it requires a `postgresql` TypeInstance of Type [`cap.type.database.postgresql.config`](och-content/type/database/postgresql/config.yaml) and input parameters [`cap.type.database.postgresql.database-input`](och-content/type/database/postgresql/database-input.yaml), and outputs a `database` TypeInstance of Type [`cap.type.database.postgresql.database`](och-content/type/database/postgresql/database.yaml). In the step, we are providing the inputs to the Interface via the `.arguments.artifacts` field. We also have to map the output of this step to our output definitions in `additionalOutput` and the implemented Interface in the `voltron-outputTypeInstances` field.

The `render-helm-args` and `fill-params-in-helm-args` steps are used to prepare the input parameters for the `helm.run` Interface. Jinja templating is used here to render the Helm runner arguments with the required data from the `postgresql` and `database` TypeInstances. Those steps don't create any TypeInstances are serve only the purpose of creating the input parameters for the Helm runner.

TODO: where to find documentation on specific runners, i.e. input parameters for them?

> To create the input parameters for `helm.run` we have to use data in two artifacts. As the current `jinja.run` Interface consumes only a template and a single variables input, we have to perform this operation twice. To seperate the variables substituted in the first and second operation, we escape some parts using `{% raw %} ... {% endraw %}`, which is removed in the first templating operation and will be processed in the second operation.
>
> In the future we might improve the ways, on how to process artifacts in the workflow.

The last step launches the Helm runner, deploys the Confluence server and creates the `confluence-config` and `confluence-helm-release` TypeInstances. The `confluence-config` TypeInstance data was provided by the Helm runner in the `additional` output artifacts from this step. Check the Helm runner documentation, on how the `additional` output is created.

## Populate the manifests into OCH

After we have the manifests ready, we can start our local Voltron environment. In the root of the cloned `go-voltron` repository run:
```
ENABLE_POPULATOR=false make dev-cluster
```

This can take a few minutes. We disabled the populator sidecar in OCH public, as we will populate the data from our local repository using the populator.

To populate the data, you will need to first setup port-forwarding to the Neo4j database service:
```
kubectl port-forward -n neo4j svc/neo4j-neo4j 7474 7687
```

Then populate the data, with the populator:
```
APP_JSONPUBLISHADDR=<your-local-docker-ip-address> APP_MANIFESTS_PATH=och-content ./populator .

APP_JSONPUBLISHADDR=http://172.17.0.1 APP_MANIFESTS_PATH=och-content populator .
```

## Run your new action

Now we will create the action, to trigger the Confluence installation. Open `https://gateway.voltron.local/` in your browser.
Then copy the following queries, variables and HTTP headers to the GraphQL playground:

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

```json
{
  "Authorization": "Basic Z3JhcGhxbDp0MHBfczNjcjN0"
}
```


Execute the `CreateAction` mutation. This will create the Action resource in Voltron. Wait till the Action is in the `READY_TO_RUN` phase. You can use the `GetAction` query to check the phase of your Action.

After it is in the `READY_TO_RUN` phase, you can see the workflow, which will be execute in the `renderedAction` field. To run the Action, execute the `RunAction` mutation.
