## Composed Action

The composed Action refers to a syntax sugar where you can tell Capact Engine to run a given set of Action and monitor them till all are finished.

To run a single Action you need to:
1. Create Action:
    ```
    capact action create --name psql-install cap.interface.database.postgresql.install --parameters-from-file /tmp/psql-install.yaml
    ```

4. Wait unit it's rendered:
    ```
    capact action get psql-install
    ```

7. Run it:
    ```
    capact action run psql-install
    ```

10. Wait until it's finished:
    ```
    capact action watch psql-install
    ```

13. Extract the output TypeInstances:
    ```
    capact action get psql-install -ojson | jq -r '.Actions[].output.typeInstances | map(select(.typeRef.path == "cap.type.database.postgresql.config"))'
    ```


And if you want to use the created PostgreSQL database for other Action, you need to repeat all those steps manually. For now, we overcome this issue by creating Implementation "umbrella workflows". This, unfortunately, results in an additional boilerplate and cause that steps are not atomic causing problem such as https://github.com/capactio/capact/issues/695 or https://github.com/capactio/capact/issues/563. Additionally, such approach make the individual Action less reusable as they don't have a single responsibility.

### ComposedAction

The main responsibility of the Composed Action CR is to schedule a set of Action, watch their execution and pass the input parameters between the steps.

```yaml
apiVersion: core.capact.io/v1alpha1
kind: ComposedAction
metadata:
  name: "full-spec"
spec:
  steps:
    psql:
      interface:
        path: "cap.interface.database.postgresql.install"
      input:
        input-parameters:
          raw:
            data: |
              superuser:
                username: superuser
              defaultDBName: postgres
    create-user:
    	# Alternative name: dependencies, depends-on or dependsOn or needs
      depends: ["psql"] # this can be removed, we can create own DAG based on the inputs.

      interface:
        path: "cap.interface.database.postgresql.create-user"
      input:
       postgresql:
         from: psql.postgresql
       input-parameters:
         raw:
           data: |
             name: mattermost

    create-db:
      # Alternative name: dependencies, depends-on or dependsOn or needs
      depends: ["create-user"]

      interface:
        path: "cap.interface.database.postgresql.create-user"
      input:
        postgresql:
          from: psql.postgresql
        input-parameters:
          raw:
            data: |
              name: mattermost
              owner: mattermost
    install:
      # Alternative name: dependencies, depends-on or dependsOn or needs
      depends: ["create-db"]

      interface:
        path: "cap.interface.productivity.mattermost.install"
      input:
        postgresql:
          from: psql.postgresql
        database:
          from: create-db.database
        user:
          from: create-user.user
  advancedRendering:
    enabled: false
  run: false
  dryRun: false
  cancel: false
```

In the future, we can also add a native support for direct Implementation references in steps. When users compose an Action, they already know what they want to achieve. Instead of creating a complex Action Policy, specific Implementations can be selected with an implementation-specific input attached to them. In this way, the Interfaces would be a nice grouping feature used on UI and later user by browsing a catalog of available service, can compose a given Action. Such Action will be easier to rationalize and requires less cognitive overhead as it describes in a declarative way want will be done.

Still, the Global Policy can be used in this approach. When user browse a catalog of available service, only those which conforms with a global policy are displayed.

### Persistence

It could be useful to share a ComposedAction between users. To do so, we can add a new entity under Public Hub tree called `template`:

```
└── manifests
   ├── attribute
   ├── core
   ├── implementation
   ├── interface
   ├── template # holds templates
   ├── type
   └── vendor
```

The new `Template` kind will hold any data indexed by its kind. For example:

```yaml
ocfVersion: 0.0.1
revision: 0.1.0
kind: Template
metadata:
  name: config
  prefix: cap.type.analytics.elasticsearch
spec:
  kind: ComposedAction # indexable
  value:
    apiVersion: core.capact.io/v1alpha1
    kind: ComposedAction
    metadata:
      name: "full-spec"
    spec:
      steps:
        # ...trimmed...
```

In the future, this can also be used to share other manifest, where dedicated Implementation and Interface are simply overhead.

## Implementation

The implementation of the Composed Action can be split into two stages.

**MVP scope**
- Implement CR ComposedAction Operator:
	- use [Kubebuilder](https://book.kubebuilder.io/) to bootstrap a new project,
	- update CI,
	- create a new Helm chart,
	- ignore the `depends` property and run all steps in sequence.
	- add minimal E2E tests coverage.
- Update CLI to support running ComposedAction CR,
- Document a new feature on website.

**Follow-ups**
- Add support for parallel and sequence execution,
- Add support for error handling and retries,
- Add support to store templates in Public Hub,
- Add support for direct Implementation references in steps,
- Add support to create ComposedAction via Dashboard:
  - Later a dedicated block builder can be added.

### Decision

During the investigation, I didn't spot any issue with the ComposeAction implementation. It looks to be straight forward, however there are multiple areas that need to be covered even in the minimal scope. However, the main inspire for implementing this feature was the [Support dynamic TypeInstances in umbrella workflow](https://github.com/capactio/capact/issues/695) issue. Fortunately, we found out an easier way to overcome it. For now, we decided to postpone it as the minimal implementation would take about 8MD where the proposed alternative needs only 3MD. This feature is likely to be implemented in the future, as it also solve a few more issues.
