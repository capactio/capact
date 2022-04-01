## Composed Action

- Should it mimic the Implementation manifest functionality?
- Can we only support implementation path directly?



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
