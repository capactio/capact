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
- [Draft](#draft)
  * [User Stories](#user-stories)
    + [Identify Action behavior (create/delete/upsert/update/get/list)](#identify-action-behavior-createdeleteupsertupdategetlist)
- [Consequences](#consequences)
- [Alternatives](#alternatives)

<!-- tocstop -->

Motivation
----------

The Voltron projects enables users easily define Actions that depends on generic capabilities instead of hard dependencies. By doing so, we can build multi-cloud, portable solutions. All Actions should work on defined Types. A given Action can consume or/and produce artifact(s). For that purpose we introduced the TypeInstance entity which is stored in Local OCH.

Currently, we struggle with defining flow for passing, creating and deleting the TypeInstances. As a result, we cannot estimate work for artifacts implementation.  

### Goal

-	Define how to identify Action behavior so we know if it creates/deletes/upserts/updates/gets/lists TypeInstances.
-	Define how to populate Action Workflow with input TypeInstances.
-	Define how to upload the generated TypeInstance from Action workflow to Local OCH.
-	Define how Action Workflow can delete the TypeInstance from Local OCH.
-	Define how to specify relations between generated TypeInstances.
-	Define where the required and optional input TypeInstances should be defined. On Interface or on Implementation?
-	Define how to override optional input TypeInstances. For example, pass already existing database.

### Non-goal

-	Provide a working POC. Currently, we are in the early stage and providing POC is to complex as we do not have implemented the base logic.  
-	Define how to store the TypeInstance in Local OCH with the preservation of Type composition.

Proposal
--------

General notes:

1.	The required input TypeInstances (artifacts) are defined on Interfaces only.
2.	The optional input TypeInstances are defined on Implementation only.
3.	The input and output TypeInstances always have a name. This solves problem when there are multiple input/output TypeInstances which refer to the same Type (e.g. backup and main database)  
4.	Action can produce only TypeInstances. We don't support dedicated user info (e.g. similar to _NOTES.txt from Helm), for Alpha and GA.

Draft
-----

```yaml
ocfVersion: 0.0.1
revision: 0.1.2
kind: Interface
metadata:
  prefix: cap.interface.cms.wordpress
  name: install
  description: WordPress installation
spec:
  input: 
    jsonSchema: # input schema, holds information that can be specified by user e.g. db size, name etc. 
      ref: cap.type.cms.wordpress.install-input:1.0.1
    typeInstances: # all required (artifacts)
      # pass as an input one instance of cap.type.cms.wordpress.config Type based on ID that user should provide.
      - name: backend_db # needs to be used in Implementation
        type: cap.type.db.mysql.config
        permission: readWrite # based on that we can 
      # pass as an input all available instances of cap.type.cms.wordpress.config Type
      - name: all_available_db
        type: []cap.type.db.mysql.config
        permission: readWrite
  output:
    typeInstances: # it's an TypeInstance that is created as a result of executed action
      - name: wp_config 
        type: cap.type.cms.wordpress.config

signature:
  och: eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9
```

### User Stories

#### Identify Action behavior (create/delete/upsert/update/get/list)

**Roles**

-	Action Developer
-	Action User

**Background**

There should be an easy way to define the Action behavior. It's necessary because our Engine needs to know how to handle specified TypeInstances. Additionally, this is used on UI to filter actions which are not depended on other TypeInstances, e.g. Actions is not upgrade, delete, etc.

Identified operations:

-	Get
-	List
-	Create
-	Delete
-	Update
-	Upsert

**Suggested solution**

Use the information from the Input/Output property defined in Interface.

-	Get operation
    ```yaml
    input: 
      typeInstances:
        - name: backend_db
          type: cap.type.db.mysql.config
          permission: read 
    ```
    
-	List operation
    ```yaml
    input: 
      typeInstances:
        - name: backend_db
          type: []cap.type.db.mysql.config # <- identifies a list of objects
          permission: read 
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
    To be figure out.
    ```
-	Update operation
    ```yaml
    input: 
      typeInstances:
        - name: backend_db
          type: []cap.type.db.mysql.config # <- identifies a list of objects
          permission: read 
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

This seems to be quite verbose and increase the overall boilerplate which is already huge.



Consequences
------------

Once approved, these are the consequences:

-	remove JSON param output from Interface (in the end, Interface will have input params, input artifacts and output artifacts)

# TODO

1.	Is the implementation allowed to register more objects than those listed in the interface? How to validate that? Output artifacts are defined both on Interface and Implementation?

2.	Is the implementation allowed to register less object than those listed in the interface? How to validate that?

3.	How artifacts can be overridden in Implementation (e.g. specify already existing db) (Define how to override TypeInstances defined in Action workflow.) Define the optional artifacts on Implementation level with required aliases?

4.	Define how to upload the generated TypeInstance from Action workflow to Local OCH.

5.	Define how to specify relations between generated TypeInstances.

6.	How and when to delete the instance from local OCH?

7.	How to populate workflow with the instances from the local OCH Saving/loading artifacts - Docker images that can be used in Argo workflow as steps? Input: user-friendly output + artifacts

8.	Optional vs required artifacts Examples: Required: db config when creating new database user Optional: db config for existing database when installing Wordpress

9.	After GA: Artifacts permissions: read/write

defining relations between artifacts (Artifacts group?) is allowed in the manifests by user

nice to have: if possible, unify the artifacts definition with requires section
