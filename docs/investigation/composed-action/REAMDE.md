## Composed Action

The composed Action refers to a syntax sugar where you can tell Capact Engine to run a given set of Action and monitor them till all are finished.

To run a single Action you need to:
1. Create Action: `capact action create --name psql-install cap.interface.database.postgresql.install --parameters-from-file /tmp/psql-install.yaml`
2. Wait unit it's rendered: `capact action get psql-install`
3. Run it: `capact action run psql-install`
4. Wait until it's finished: `capact action watch psql-install`
5. Extract the output TypeInstances: `capact action get psql-install -ojson | jq -r '.Actions[].output.typeInstances | map(select(.typeRef.path == "cap.type.database.postgresql.config"))'`

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
  input:
    parameters:
      secretRef:
        name: "full-spec-params"
    typeInstances:
      foo:
        id: fee33a5e-d957-488a-86bd-5dacd4120312
      bar:
        id: 563a79eb-7417-4e11-aa4b-d93076c04e48
  advancedRendering:
    enabled: false
  run: false
  dryRun: false
  cancel: false
```

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
- Implement CR ComposedAction Operator
	- use [Kubebuilder](https://book.kubebuilder.io/) to bootstrap a new project,
	- update CI,
	- create a new Helm chart,
	- ignore the `depends` property and run all steps in sequence.
	- add minimal E2E tests coverage.
- Update CLI to support running ComposedAction CR
- Document a new feature on website

**Follow-ups**
- Add support for parallel and sequence execution,
- Add support for error handling and retries,
- Add support to store templates in Public Hub,
- Add support to create ComposedAction via Dashboard:
  - Later a dedicated block builder can be added.

### Summary

During the investigation, I didn't spot any issue with the ComposeAction implementation. It looks to be straight forward, however there are multiple areas that need to be covered even in the minimal scope.
