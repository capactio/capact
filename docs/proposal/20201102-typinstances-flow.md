#  Handling TypeInstances in Interfaces and Implementations

Created on 2020-11-02 by Mateusz Szostok ([@mszostok](https://github.com/mszostok/))

##  Overview

This document describes the approach for handling the TypeInstances (artifacts) between different Voltron components.

<!-- toc -->

- [Motivation](#motivation)
  * [Goal](#goal)
  * [Non-goal](#non-goal)
- [Proposal](#proposal)
  * [Required input and output TypeInstances](#required-input-and-output-typeinstances)
    + [Suggested solution](#suggested-solution)
    + [Alternatives](#alternatives)
  * [Handle optional input TypeInstances](#handle-optional-input-typeinstances)
    + [Suggested solution](#suggested-solution-1)
    + [Alternatives](#alternatives-1)
  * [Identify Action behavior (create/delete/upsert/update/get/list)](#identify-action-behavior-createdeleteupsertupdategetlist)
    + [Suggested solution](#suggested-solution-2)
    + [Alternatives](#alternatives-2)
  * [Additional output TypeInstances and relations between them](#additional-output-typeinstances-and-relations-between-them)
    + [Suggested solution](#suggested-solution-3)
  * [Populate an Action with the input TypeInstances](#populate-an-action-with-the-input-typeinstances)
    + [Suggested solution](#suggested-solution-4)
  * [Upload Action artifacts to Local OCH](#upload-action-artifacts-to-local-och)
    + [Suggested solution](#suggested-solution-5)
    + [Alternatives](#alternatives-3)
  * [Delete the TypeInstance from Local OCH by Action](#delete-the-typeinstance-from-local-och-by-action)
    + [Suggested solution](#suggested-solution-6)
- [Consequences](#consequences)

<!-- tocstop -->

##  Motivation

The Voltron project enables users to easily define Actions that depend on generic capabilities instead of hard dependencies. By doing so, we can build multi-cloud, portable solutions.

All Actions should work on defined Types. A given Action can consume or/and produce artifact(s). For that purpose, we introduced the TypeInstance entity which is stored in Local OCH.

Currently, we are struggling with defining the flow for passing, creating, and deleting TypeInstances. As a result, we cannot estimate the work for artifacts implementation.

###  Goal

-	[Define how to handle required input TypeInstances.](#required-and-optional-input-typeinstances)
-	[Define how to handle optional input TypeInstances. For example, pass an already existing database.](#handle-optional-input-typeinstances)
-	[Define how to identify Action behavior so we know if it creates/deletes/upserts/updates/gets/lists TypeInstances.](#identify-action-behavior-createdeleteupsertupdategetlist)
-	[Define how to populate Action with input TypeInstances.](#populate-an-action-with-the-input-typeinstances)
-	[Define how to upload the generated artifacts from Action workflow to Local OCH.](#upload-action-artifacts-to-local-och)
-	[Define how Action can delete the TypeInstance from Local OCH.](#delete-the-typeinstance-from-local-och-by-action)
-	[Define how to specify additional output TypeInstances and relations between them.](#additional-output-typeinstances-and-relations-between-them)

###  Non-goal

-	Provide a working POC. Currently, we are in the early stage, and providing POC is too complex as we do not have implemented the base logic.  
-	Define how to store the TypeInstance in Local OCH with the preservation of Type composition.
-	Define the final syntax for Action Workflow. This will be done in a separate task by taking into account the Argo Workflow syntax.  

##  Proposal

Terminology

| Term             | Definition                                                                                                                        |
|------------------|-----------------------------------------------------------------------------------------------------------------------------------|
| Artifacts        | Input/Output object returned by steps in a given workflow.                                                                        |
| TypeInstance     | An instance of a given Type that is stored in the Local OCH. Artifacts uploaded from workflow to Local OCH becomes TypeInstances. |
| Action developer | Person who defines the `action` property in Implementation manifest.                                                              |

General notes

1.	The required input TypeInstances are defined on Interfaces only.
2.	The optional input TypeInstances are defined on Implementation only.
3.	The input and output TypeInstances always have a name. This solves the problem when there are multiple input/output TypeInstances which refer to the same Type (e.g. backup and main database)  
4.	Action can produce only TypeInstances. We don't support output parameters, outputted for user (e.g. similar to `NOTES.txt` from Helm). We can revisit it after GA as this won't be a breaking change.

###  Required input and output TypeInstances

Actors

-	Action developer

####  Suggested solution

We have the Interface entity which defines the input and output parameters. To fulfill a given Interface, Implementation needs to accept the same input and returns the same output parameters.

The **required** input and output TypeInstances should be defined on Interface. By doing so, we can ensure that Implementations are exchangeable and do not introduce new requirements.

For the Beta and GA only one TypeInstance can be declared as the output. As a result we can simplify implementation for defining [relations between generated artifacts](#specify-relations-between-generated-artifacts).

<details> <summary>Example</summary>

Syntax for Interface:

```yaml
kind: Interface
metadata:
  prefix: cap.interface.cms.wordpress
  name: install
spec:
  input: 
    parameters: # holds information that can be specified by user e.g. db size, name etc. 
      jsonSchema: |-
        {
          "type": "object",
          "properties": {
            "name": {
              "type": "string"
            }
          }
        }
    typeInstances: # all bellow entities are required and need to be passed to Implementation
      # pass as an input one instance of cap.type.cms.wordpress.config Type based on ID that user should provide.
      backend_db: # unique name that needs to be used in Implementation
        typeRef:
          path: cap.type.db.mysql.config
          revision: 0.1.0
        verbs: ["get", "update"]
  output:
    typeInstances: # it's an TypeInstance that is created as a result of executed action. ONLY one can be declared for Beta and GA.
      wp_config: 
        typeRef:
          path: cap.type.cms.wordpress.config
          revision: 0.1.0
```

</details>

####  Alternatives

There is an option to define `typeInstances` as a list instead of a map.

<details> <summary>Example</summary>

```yaml
    typeInstances:
      - name: backend_db
        typeRef:
          path: cap.type.db.mysql.config
          revision: 0.1.0
        verbs: ["get", "update"]
```

</details>

Unfortunately, in that way, we cannot easily enforce that the names won't be repeated, and we cannot benefit from native YAML syntax support.

###  Handle optional input TypeInstances

Actors

-	Action developer

Only Implementation knows that something can be swapped out e.g. defined workflow can handle the situation when a user passes the existing database and reuse it instead of creating a new one. As a result, the **optional** input TypeInstances should be defined on Implementation.

Specifying optional input TypeInstance on Implementation, cause that user is able to discover and pass optional TypeInstance only during the render process of a specific Implementation.

We also need to take into account that the Action developer should be able to handle optional TypeInstances and if a given TypeInstance is available then skip a given step(s).

####  Suggested solution

Introduce the new property `input.additionalInput.typeInstances` and template language that can be used by Action workflow developers. The Voltron Engine needs to handle optional TypeInstances and render the final manifest based on the template syntax.

<details> <summary>Example</summary>

Implementation workflow:

```yaml
kind: Implementation
# ...
spec:
  # Workflow Developer needs to specify optional input TypeInstances 
  additionalInput:
    # maybe some day also parameters: {}
    typeInstances: # names need to be different from those define under spec.input.typeInstances in Interface
      mysql-config: 
        typeRef:
          path: cap.type.db.mysql.config
          revision: 0.1.0
        verbs: ["get", "update"]
  
  action:
    type: argo.run
    args:
      workflow:
        steps:
          {{ if additionalInput.typeInstances.mysql-config == nil }}
          - name: gcp-create-service-account
            outputs:
              artifacts:
                - name: "gcp-sa"
          - name: create-cloud-sql
            inputs:
              artifacts:
                - name: "gcp-sa"
            outputs:
              artifacts:
                - name: "mysql-config"
          {{ endif }}
          - name: mysql-create-db
            inputs:
              artifacts:
                - name: "mysql-config"
```

Rendered Implementation workflow when `additionalInput.typeInstances.mysql-config` was specified by user:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: mysql-create-db-
spec:
  entrypoint: work
  templates:
    - name: downloads-instances
      container:
        image: gcr.io/projectvoltron/type-instance-fetcher:0.0.1
      outputs:
        # export a global artifact. The artifact will be programatically available in the completed
        # workflow object under: workflow.outputs.artifacts
        # globalName corresponds to the name defined in Interface `spec.input.typeInstances` section.
        artifacts:
        - name: mysql-config
          globalName: mysql-config
  
    - name: mysql-create-db
      container:
        image: gcr.io/projectvoltron/actions/mysql-create-db:0.0.1
      arguments:
        artifacts:
        - name: mysql-config
          from: "{{workflow.outputs.artifacts.mysql-config}}"
```

Rendered Implementation workflow when `additionalInput.typeInstances.mysql-config` wasn't specified by user:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: mysql-create-db-
spec:
  entrypoint: work
  templates:
    - name: gcp-create-service-account
      container:
        image: gcr.io/google-och/actions/gcp-create-service-account:0.0.1
      outputs:
        artifacts:
        - name: gcp-sa
    - name: create-cloud-sql
      container:
        image: gcr.io/google-och/actions/create-cloud-sql:0.0.1
      arguments:
        artifacts:
        - name: mysql-config
          from: "{{steps.gcp-create-service-account.outputs.artifacts.gcp-sa}}"
      outputs:
        artifacts:
        - name: mysql-config
          # Global as this is mentioned as optional input, so it can be also populated from Local OCH by initial step
          globalName: mysql-config
    - name: mysql-create-db
      container:
        image: gcr.io/projectvoltron/actions/mysql-create-db:0.0.1
      arguments:
        artifacts:
        - name: mysql-config
          from: "{{workflow.outputs.artifacts.mysql-config}}"
```

</details>

####  Alternatives

Instead of giving the Action developer the option to use the template language we can determine that directly during the render process. The step could be automatically removed if the artifact name specified as an output of the given step matches with the one which was passed to the Action. Unfortunately, this solution hides a lot and does not support a more complex scenario e.g. a given step outputs more that one artifact or action developer wants to remove more steps when a given TypeInstance was passed.

<details> <summary>Example</summary>

Implementation workflow:

```yaml
action:
  type: argo.run
  args:
    workflow:
      steps:
      - name: gcp-create-service-account
        outputs:
          artifacts:
            - name: "gcp-sa"
      - name: create-cloud-sql
        inputs:
          artifacts:
            - name: "gcp-sa"
        outputs:
          artifacts:
            - name: "mysql-config"
      - name: mysql-create-db
        inputs:
          artifacts:
            - name: "mysql-config"
```

Rendered Implementation workflow when `input.typeInstances.mysql-config` was passed by user:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: mysql-create-db-
spec:
  entrypoint: work
  templates:
    - name: downloads-instances
      container:
        image: gcr.io/projectvoltron/type-instance-fetcher:0.0.1
      outputs:
        artifacts:
        - name: mysql-config
          path: /tmp/mysql-config
          globalName: mysql-config
  
    - name: mysql-create-db
      container:
        image: gcr.io/projectvoltron/actions/mysql-create-db:0.0.1
      arguments:
        artifacts:
        - name: mysql-config
          from: "{{workflow.outputs.artifacts.mysql-config}}"
```

Engine removes automatically the step which produces the passed TypeInstance and automatically removes not needed steps such as `gcp-create-service-account` because the only consumer of its output TypeInstance was removed. This logic should take into account that some steps can produce TypeInstance which doesn't have to be consumed e.g. generates a report. In such a case it should be garbage collected.

</details>

###  Identify Action behavior (create/delete/upsert/update/get/list)

Actors

-	Action developer
-	Action user

There should be an easy way to define Action behavior. It's necessary because our Engine needs to know how to handle specified TypeInstances. Additionally, this is used on UI to filter actions that are not dependent on other TypeInstances, e.g. Actions is not `upgrade`, `delete`, etc.

Identified operations:

-	Get
-	List
-	Create
-	Delete
-	Update
-	Upsert

####  Suggested solution

Use the information from the `input`/`output` property defined in Interface. For each TypeInstance we can define `verbs` property to specify what kind of operation will be executed against that TypeInstance. Based on that we can determine Action behavior.

| Verbs    | Description                                                                               |
|----------|-------------------------------------------------------------------------------------------|
| `get`    | Specify that the input is a single TypeInstance that is in read-only mode.                |
| `list`   | Specify that the input is a list of all TypeInstances from Local OCH in read-only mode.   |
| `create` | This is automatically set for output TypeInstances. Core Action stores them in Local OCH. |
| `update` | Specify that the input TypeInstance is modified in Action.                                |
| `delete` | Specify that the input TypeInstance is deleted by Action.                                 |

<details> <summary>Example</summary>

-	Get operation

	```yaml
	input: 
	  typeInstances:
	    backend_db:
	      typeRef:
	        path: cap.type.db.mysql.config
	        revision: 0.1.0
	      verbs: ["get"] 
	```

-	List operation

	```yaml
	input: 
	  typeInstances:
	    backend_db:
	      typeRef:
	        path: cap.type.db.mysql.config
	        revision: 0.1.0
	      verbs: ["list"] 
	```

-	Create operation

	```yaml
	output: 
	  typeInstances:
	    wp_config:
	      typeRef:
            path: cap.type.cms.wordpress.config
            revision: 0.1.0
	```

-	Delete operation

	```yaml
	input: 
	  typeInstances:
	    backend_db:
	      typeRef:
	        path: cap.type.db.mysql.config
	        revision: 0.1.0
	      verbs: ["delete"] 
	```

-	Update operation

	```yaml
	input: 
	  typeInstances:
	    backend_db:
	      typeRef:
	        path: cap.type.db.mysql.config
	        revision: 0.1.0
	      verbs: ["get", "update"] 
	```

</details>

####  Alternatives

1.	We could introduce and use Attributes defined on Interfaces to explicitly mark its behavior.

	<details> <summary>Details</summary>

	```yaml
	kind: Interface
	metadata:
	  prefix: cap.interface.cms.wordpress
	  name: install
	  tags:
	    cap.core.action.install: true # cap.core.action.upgrade |  cap.core.action.upsert | cap.core.action.uninstall
	# ...
	```

	This seems to be simpler and more explicit but at the same time, it is redundant, as we need to define that also per TypeInstances.

	</details>

2.	We can introduce dedicated Action types.

	<details> <summary>Details</summary>

	```yaml
	kind: Interface
	metadata:
	  prefix: cap.interface.cms.wordpress
	  name: install
	spec:
	  input:
	    typeInstances:
	      cap.core.action.register: # such instance is produced
	        - type: cap.type.cms.wordpress.config
	      cap.core.action.modify: # modifies existing instance
	        - type: cap.type.database.mysql.config
	      cap.core.action.list: #  pass the list of all existing instance
	        - type: cap.type.database.mysql.config
	      cap.core.action.get: #  pass the existing instance by ID
	        - type: cap.type.database.mysql.config
	# ...
	```

	This seems to be quite verbose and increases the overall boilerplate which is already huge.

	</details>

3.	We can use `permissions` instead of `verbs`.

	<details> <summary>Details</summary>

	```yaml
	input: 
	  typeInstances:
	    - name: backend_db
	      type: cap.type.db.mysql.config
	      permissions: ["read"] 
	```

	With the `permissions` name, it's not so easy to explain that it affects the rendered workflow, and e.g. based on that Engine is injecting some steps automatically as describe [here](#populate-an-action-with-the-input-typeinstances).

	</details>

###  Additional output TypeInstances and relations between them

Actors

-	Action developer

If we know the relations between the TypeInstance e.g. that Jira instance uses a given database we can easily show the graph with those relations and based on that user can detect dependencies and also check if downtime of a given component can affect other parts of the system.

####  Suggested solution

Specify the relations between the TypeInstances in the Implementation. As describe in [upload artifacts](#upload-action-artifacts-to-local-och) section, in the Implementation, the Action developer knows all details about all output artifacts and how they are related to each other. For now, the Interface can define only one TypeInstance in the output section, so the relations property is not necessary on the Interface type. As a result, it simplifies the solution for Beta and GA as we don't need to take care of proper merging those relations defined on Interface and Implementation.

This property is required if the Implementation wants to upload more than one artifact. By doing so, in the future we can implement a more sophisticated mechanism for deleting the TypeInstances as we will know relations between them, e.g. when someone will schedule removing the WordPress Config TypeInstance we will know that the correlated Ingress TypeInstance also should be removed.

<details> <summary>Example</summary>

```yaml
kind: Implementation
# ...
spec:
  # when saving artifact in Local OCH we can create a proper edges.
  additionalOutput:
    typeInstances: # list all optional artifacts with type references
      mysql_config:
        typeRef:
          path: cap.type.db.mysql.config
          revision: 0.1.0	     
      ingress_config:
        typeRef:
          path: cap.type.networking.ingress.config
          revision: 0.1.0
      cloudsql_config:
        typeRef:
          path: cap.type.gcp.cloudsql.config
          revision: 0.1.0   
    typeInstanceRelations:
      wordpress_config: # artifact name
        uses: # names of all artifacts that WP config depends on
          - cloudsql_config
          - ingress_config
      cloudsql_config: # artifact name
        uses:
          - mysql_config
```

</details>

###  Populate an Action with the input TypeInstances

Actors

-	Voltron Engine

####  Suggested solution

If Action requires input TypeInstance, the Voltron Engine adds an initial download step to the Workflow. This step runs the core Action which connects to Local OCH and downloads TypeInstances and exposes them as a [global Argo artifacts](https://github.com/argoproj/argo/blob/6016ebdd94115ae3fb13cadbecd27cf2bc390657/examples/global-outputs.yaml#L33-L36), so they can be accessed by other steps via `{{workflow.outputs.artifacts.<name>}}`.

The global Argo artifacts seem to be the only possible solution as the steps output artifacts are scoped to a given template. This assumption is based on [argo-workflows investigation](../investigation/argo-workflows/README.md) document.

<details> <summary>Example</summary>

Interface:

```yaml
kind: Interface
metadata:
  prefix: cap.interface.db.mysql
  name: create-db
spec:
  input: 
    typeInstances:
      mysql-config:
	    typeRef:
	      path: cap.type.db.mysql.config
	      revision: 0.1.0
        verbs: ["get"]
```

Implementation workflow:

```yaml
action:
type: argo.run
args:
  workflow:
    steps:
      - name: mysql-create-db
        inputs:
          artifacts:
            - name: "mysql-config"
```

Rendered Implementation workflow:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: mysql-create-db-
spec:
  entrypoint: work
  templates:
    - name: downloads-instances
      container:
        image: gcr.io/projectvoltron/type-instance-fetcher:0.0.1
      outputs:
        # export a global artifact. The artifact will be programatically available in the completed
        # workflow object under: workflow.outputs.artifacts
        # globalName corresponds to the name defined in Interface `spec.input.typeInstances` section.
        artifacts:
        - name: mysql-config
          path: /tmp/mysql-config
          globalName: mysql-config
  
    - name: mysql-create-db
      container:
        image: gcr.io/projectvoltron/actions/mysql-create-db:0.0.1
      arguments:
        artifacts:
        - name: mysql-config
          from: "{{workflow.outputs.artifacts.mysql-config}}"
```

</details>

###  Upload Action artifacts to Local OCH

Actors

-	Action developer

####  Suggested solution

The Voltron Engine could automatically add a step at the end of the Workflow to uploads all TypeInstances specified under `spec.output`. Unfortunately, it gets complicated when the Action developer wants to upload additional TypeInstances. To solve that problem, the Action developer needs to describe the relations between additional output TypeInstances as described in [this](#additional-output-typeinstances-and-relations-between-them) section. In that way, the Voltron Engine can add a step that can upload artifacts automatically.

To update TypeInstance in a workflow, the workflow output artifact name should match the workflow input artifact. The upload step overwrites TypeInstance using PUT-like operation with a proper `resourceVersion`. If there is a conflict, the workflow fails and can be retried.

> **NOTE**: The ability to update the TypeInstance is planned for GA.

Restrictions:

-	Implementation MUST upload all TypeInstances which are defined under the `spec.output` property in Interface. Uploaded TypeInstances MUST be exactly the same as those defined in Interface or being an extension thereof.

-	Implementation is allowed to upload more TypeInstances than those listed in the Interface. To do so, Action developer needs to describe the relations between additional output TypeInstances as described in [this](#additional-output-typeinstances-and-relations-between-them) section.

> **NOTE:** For the Beta and GA Engine doesn't validate above restrictions.

<details> <summary>Example</summary>

Interface:

```yaml
kind: Interface
metadata:
  prefix: cap.interface.management.jira
  name: install
spec:
  output: 
    typeInstances:
      jira_config:
        type: cap.type.management.jira.config
```

Implementation workflow:

```yaml
kind: Implementation
# ...
spec:
  # Workflow developer needs to specify relations between additional output.
  # We use that to create a proper edges when saving artifact in Local OCH.
  additionalOutput:
    typeInstances: # list all optional artifacts with type references
      mysql_config:
        typeRef:
          path: cap.type.db.mysql.config
          revision: 0.1.0	     
      ingress_config:
        typeRef:
          path: cap.type.networking.ingress.config
          revision: 0.1.0  
    typeInstanceRelations:
      jira_config:
        uses:
          - mysql_config
          - ingress_config
  action:
    type: argo.run
    args:
      workflow:
        steps:
          - name: create-mysql-db
            {{ actionFrom: cap.interface.db.mysql.install }}
            outputs:
              artifacts:
                - name: "mysql_config"
          - name: install-jira
            {{ actionFrom: cap.interfaces.management.jira.install }}
            inputs:
              artifacts:
                - name: "mysql_config"
            outputs:
              artifacts:
                - name: "jira_config"
          - name: expose-ingress
            {{ actionFrom: cap.interfaces.gcp.create-cloud-sql }}
            inputs:
              artifacts:
                - name: "jira_config"
            outputs:
              artifacts:
                - name: "ingress_config"
```

</details>

####  Alternatives

The Action Developer is able to use core upload Action and define manually which TypeInstances should be uploaded.

<details> <summary>Example</summary>

Interface:

```yaml
kind: Interface
metadata:
  prefix: cap.interface.db.mysql
  name: install
spec:
  output: 
    typeInstances:
      mysql-config:
        typeRef:
          path: cap.type.db.mysql.config
          revision: 0.1.0
```

Implementation workflow:

```yaml
kind: Implementation
# ...
spec:
  action:
  type: argo.run
  args:
    workflow:
      steps:
        - name: gcp-create-service-account
          {{ actionFrom: cap.interfaces.gcp.create-service-account }}
          outputs:
            artifacts:
              - name: "gcp-sa"
        - name: create-cloud-sql
          {{ actionFrom: cap.interfaces.gcp.create-cloud-sql }}
          inputs:
            artifacts:
              - name: "gcp-sa"
          outputs:
            artifacts:
              - name: "mysql-config"
```

This solution was rejected as we found out that we can do it also automatically. Additional, it is similar to the populate and deletes TypeInstance actions which are also automatically injected by Voltron Engine.

</details>

###  Delete the TypeInstance from Local OCH by Action

Actors

-	Voltron Engine

####  Suggested solution

Based on the suggestion from [this section](#populate-an-action-with-the-input-typeinstances). The Voltron Engine adds an initial download step to the Workflow to download all TypeInstances specified under `spec.input` which have `read` verb. In the same way, we can handle the deletion of the TypeInstances. The Voltron Engine adds a step at the end of the Workflow to delete all TypeInstances specified under `spec.input` which have a `delete` verb.

<details> <summary>Example</summary>

Interface:

```yaml
kind: Interface
metadata:
  prefix: cap.interface.db.mysql
  name: uninstall
spec:
  input: 
    typeInstances:
      mysql-config:
        typeRef:
          path: cap.type.db.mysql.config
          revision: 0.1.0
        verbs: ["get", "delete"]
```

Implementation workflow:

```yaml
action:
type: argo.run
args:
  workflow:
    steps:
      - name: mysql-delete-db
        inputs:
          artifacts:
            - name: "mysql-config"
```

Rendered Implementation workflow:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: mysql-create-db-
spec:
  entrypoint: work
  templates:
    # Download artifacts from Local OCH
    - name: downloads-instances
      container:
        image: gcr.io/projectvoltron/type-instance-fetcher:0.0.1
      outputs:
        artifacts:
        - name: mysql-config
          path: /tmp/mysql-config
          globalName: mysql-config
    # Steps from Implementation workflow
    - name: mysql-delete-db
      container:
        image: gcr.io/projectvoltron/actions/mysql-create-db:0.0.1
      arguments:
        artifacts:
        - name: mysql-config
          from: "{{workflow.outputs.artifacts.mysql-config}}"
    # Deletes artifacts from Local OCH 
    - name: delete-instances
      container:
        image: gcr.io/projectvoltron/type-instance-deleter:0.0.1
      arguments:
        artifacts:
        - name: mysql-config
          from: "{{workflow.outputs.artifacts.mysql-config}}"
```

</details>

##  Consequences

Once approved, these are the consequences:

-	Remove JSON output from Interface.
-	Rename the `input.jsonSchema` on Interface to `input.parameters.jsonSchema`.
-	Update the [OCF JSONSchemas](../../ocf-spec/0.0.1/schema) with accepted new syntax.
-	Update the GraphQL queries for Engine and OCH.
