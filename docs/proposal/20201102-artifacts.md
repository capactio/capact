Artifacts
=========

Created on 2020-11-02 by Mateusz Szostok ([@mszostok](https://github.com/mszostok/))

Overview
--------

This document describes the approach for handling the TypeInstances (artifacts) between different Voltron components.

<!-- toc -->

- [Motivation](#motivation)
  * [Goal](#goal)
  * [Non-goal](#non-goal)
- [Proposal](#proposal)
  * [Use cases](#use-cases)
    + [Required and optional input TypeInstances](#required-and-optional-input-typeinstances)
    + [Identify Action behavior (create/delete/upsert/update/get/list)](#identify-action-behavior-createdeleteupsertupdategetlist)
    + [Populate an Action with the input TypeInstances](#populate-an-action-with-the-input-typeinstances)
    + [Handle optional input TypeInstances](#handle-optional-input-typeinstances)
  * [Consequences](#consequences)
- [TODO](#todo)

<!-- tocstop -->

Motivation
----------

The Voltron projects enable users to easily define Actions that depend on generic capabilities instead of hard dependencies. By doing so, we can build multi-cloud, portable solutions. All Actions should work on defined Types. A given Action can consume or/and produce artifact(s). For that purpose, we introduced the TypeInstance entity which is stored in Local OCH.

Currently, we struggle with defining flow for passing, creating, and deleting the TypeInstances. As a result, we cannot estimate work for artifacts implementation.

### Goal

-	[Define how to identify Action behavior so we know if it creates/deletes/upserts/updates/gets/lists TypeInstances.](#identify-action-behavior-createdeleteupsertupdategetlist)
-	[Define how to populate Action with input TypeInstances.](#populate-an-action-with-the-input-typeinstances)
-	Define how to upload the generated TypeInstance from Action workflow to Local OCH.
-	Define how Action can delete the TypeInstance from Local OCH.
-	Define how to specify relations between generated TypeInstances by a given Action.
-	[Define how the required and optional input TypeInstances should be defined.](#required-and-optional-input-typeinstances)
-	[Define how to handle optional input TypeInstances. For example, pass an already existing database.](#handle-optional-input-typeinstances)

### Non-goal

-	Provide a working POC. Currently, we are in the early stage, and providing POC is too complex as we do not have implemented the base logic.  
-	Define how to store the TypeInstance in Local OCH with the preservation of Type composition.
-	Define the final syntax for Action Workflow. This will be done in a separate task by taking into account the Argo Workflow syntax.  

Proposal
--------

General notes:

1.	The required input TypeInstances (artifacts) are defined on Interfaces only.
2.	The optional input TypeInstances are defined on Implementation only.
3.	The input and output TypeInstances always have a name. This solves the problem when there are multiple input/output TypeInstances which refer to the same Type (e.g. backup and main database)  
4.	Action can produce only TypeInstances. We don't support dedicated user info (e.g. similar to _NOTES.txt from Helm), for Alpha and GA. Can be considered once again after GA as this won't be a breaking change.

### Use cases

#### Required and optional input TypeInstances

**Actors**

-	Action Developer

**Suggested solution**

We have the Interface entity which defines the input and output parameters which become Action signature.

The **required** input TypeInstances should be defined on Interface. By doing so, we can ensure that Implementations are exchangeable and do not introduce new requirements.

The **optional** input TypeInstances should be defined on Implementation. The only Implementation knows that something can be swapped out, e.g. users can pass an existing database and defined workflow can handle such a situation and reuse the given database instead of creating a new one.

Syntax for Interface:

```yaml
kind: Interface
metadata:
  prefix: cap.interface.cms.wordpress
  name: install
spec:
  input: 
    jsonSchema: # input schema, holds information that can be specified by user e.g. db size, name etc. 
      ref: cap.type.cms.wordpress.install-input:1.0.1
    typeInstances: # all bellow entities are required and need to be passed to Implementation
      # pass as an input one instance of cap.type.cms.wordpress.config Type based on ID that user should provide.
      backend_db: # unique name that needs to be used in Implementation
        type: cap.type.db.mysql.config
        permissions: ["read", "update"]
      # pass as an input all available instances of cap.type.cms.wordpress.config Type
      all_available_db:
        type: []cap.type.db.mysql.config
        permissions: ["read", "update"]
  output:
    typeInstances: # it's an TypeInstance that is created as a result of executed action
      wp_config: 
        type: cap.type.cms.wordpress.config
```

Sytanx for Implementation:

```yaml
WIP
```

**Alternatives**

There is an option to define `typeInstances` as a list instead of map:

```yaml
    typeInstances:
      - name: backend_db
        type: cap.type.db.mysql.config
        permissions: ["read", "update"] 
      - name: all_available_db
        type: []cap.type.db.mysql.config
        permissions: ["read", "update"]
```

Unfortunately, in that way, we cannot easily enforce that the names won't be repeated, and we cannot benefit from native YAML syntax support.

#### Identify Action behavior (create/delete/upsert/update/get/list)

**Actors**

-	Action Developer
-	Action User

**Background**

There should be an easy way to define Action behavior. It's necessary because our Engine needs to know how to handle specified TypeInstances. Additionally, this is used on UI to filter actions that are not dependent on other TypeInstances, e.g. Actions is not upgrade, delete, etc.

Identified operations:

-	Get
-	List
-	Create
-	Delete
-	Update
-	Upsert

**Suggested solution**

Use the information from the Input/Output property defined in Interface. Permission allows us to determine Action behavior.

| Permission | Description                                                                           |
|------------|---------------------------------------------------------------------------------------|
| `read`     | Specify that the input artifact is in read-only mode.                                 |
| `create`   | This is automatically set for output artifacts. Core Action stores them in Local OCH. |
| `update`   | Specify that the input artifact is modified in Action.                                |
| `delete`   | Specify that the input artifact is deleted by Action.                                 |

-	Get operation

	```yaml
	input: 
	  typeInstances:
	    - name: backend_db
	      type: cap.type.db.mysql.config
	      permissions: ["read"] 
	```

-	List operation

	```yaml
	input: 
	  typeInstances:
	    - name: backend_db
	      type: []cap.type.db.mysql.config # <- identifies a list of objects
	      permissions: ["read"] 
	```

-	Create operation

	```yaml
	output: 
	  typeInstances:
	    - name: wp_config
	      type: cap.type.cms.wordpress.config
	```

-	Delete operation

	```yaml
	input: 
	  typeInstances:
	    - name: backend_db
	      type: cap.type.db.mysql.config
	      permissions: ["delete"] 
	```

-	Update operation

	```yaml
	input: 
	  typeInstances:
	    - name: backend_db
	      type: cap.type.db.mysql.config
	      permissions: ["read", "update"] 
	```

-	Upsert

	```yaml
	Do we really need that?
	```

**Alternatives**

**Use Tags on Interface**

Example:

```yaml
kind: Interface
metadata:
  prefix: cap.interface.cms.wordpress
  name: install
  tags:
    cap.core.action.install: true # cap.core.action.upgrade |  cap.core.action.upsert | cap.core.action.uninstall
# ...
```

This seams to be simpler and more explicit but at the same time it is redundant, as we need to define that also per artifacts.

**Introduce dedicated Action types**

Example:

```yaml
typeInstances:
 cap.core.action.register: # such instance is produced
   - type: cap.type.cms.wordpress.config
 cap.core.action.modify: # modifies existing instance
   - type: cap.type.database.mysql.config
 cap.core.action.list: #  pass the list of all existing instance
   - type: cap.type.database.mysql.config
 cap.core.action.get: #  pass the existing instance by ID
   - type: cap.type.database.mysql.config
```

This seems to be quite verbose and increases the overall boilerplate which is already huge.

#### Populate an Action with the input TypeInstances

**Actors**

-	Voltron Engine

**Suggested solution**

If Action requires input TypeIntstance, the Voltron Engine adds an initial download step to the Workflow. This step runs the core Action which connects to Local OCH and downloads TypeInstances and exposes them as a [global Argo artifacts](https://github.com/argoproj/argo/blob/6016ebdd94115ae3fb13cadbecd27cf2bc390657/examples/global-outputs.yaml#L33-L36), so they can be accessed by other steps via `{{workflow.outputs.artifacts.db_config}}`.

The global Argo artifacts seem to be the only possible solution as the steps output artifacts are scoped to a given template. This assumption is based on [argo-workflows investigation](../investigation/argo-workflows/README.md) document.

**Implementation workflow**

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

**Rendered Implementation workflow**

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


#### Handle optional input TypeInstances

**Actor**

-	Action workflow developer

**Background**

Based on the solution from [this](#required-and-optional-input-typeinstances) section, the optional artifacts are defined on Implementation. The user is able to pass the optional TypeInstance during the render process. The workflow developer should be able to handle that situation and if a given TypeInstance is available then skip a given step(s).

**Suggested solution**

Introduce the template language that can be used by Action workflow developers. This can be resolved during the render action by Voltron Engine.

**Implementation workflow**

```yaml
action:
type: argo.run
args:
  workflow:
    steps:
      {{ if input.typeInstances.mysql-config == nil }}
      - name: mysql-install
        outputs:
          artifacts:
            - name: "mysql-config"
      {{ endif}}
      - name: mysql-create-db
        inputs:
          artifacts:
            - name: "mysql-config"
```

**Rendered Implementation workflow when `input.typeInstances.mysql-config` was passed by user**

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

**Rendered Implementation workflow when `input.typeInstances.mysql-config` wasn't passed by user**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: mysql-create-db-
spec:
  entrypoint: work
  templates:
    - name: mysql-install
      container:
        image: gcr.io/projectvoltron/actions/mysql-install:0.0.1
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

**Alternatives**

Instead of giving the Action developer the option to use the template language we can determine that directly during render action. The step could be automatically removed if the artifact name specified as an output of the action matches with the one which was passed to the Actions. Unfortunately, this solution hides a lot and does not support a more complex scenario e.g. a given step outputs more that one artifact or workflow developer wants to remove more steps when a given TypeInstance was passed.

### Consequences

Once approved, these are the consequences:

-	remove JSON output from Interface
-	update the [OCF JSONSchemas](../../ocf-spec/0.0.1/schema) with accepted new syntax  

TODO
----

1.	Is the implementation allowed to register more objects than those listed in the interface? How to validate that? Output artifacts are defined both on Interface and Implementation?

2.	Is the implementation allowed to register fewer objects than those listed in the interface? How to validate that?

3.	Is the implementation allowed to register object which extends the type defined on Interface? How to allow that? Can be validated behind the scene?

4.	How and when to delete the instance from local OCH?

5.	Define relations between artifacts (Artifacts group in the manifest?)

6.	nice to have: if possible, unify the artifacts' definition with requires section
