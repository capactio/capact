# Dynamic TypeInstances in delegated storage

## Problem

Previously, every downloaded TypeInstance in workflow had a form of an artifact with the TypeInstance value.
Now, every TypeInstance have:

```yaml
value: {} # static value, or resolved value based on context
backend:
  context: {} # additional context for storage backend
```

Jinja2 templating and Artifact Merger already support unpacking `value`.

Sometimes the artifact could contain just the `backend.context`. For example, if we took an artifact prepared for Helm Template storage:

```yaml
backend:
  context:
    chartLocation: https://charts.bitnami.com/bitnami
    driver: secrets
    name: postgresql-1648472938
    namespace: default
value: null
```

The problem is that some of further workflow steps might need actual values, and the `value` is null.

## Potential solutions


### Use `value` field to store the rendered value

- this property would always save rendered value under `.value`
  - helm runner - no need to use external backend [0.5MD]
  - in future, terraform runner would also do it in this way
- jinja2 and merger already support that
- TypeInstance download already fetches TypeInstance value (static or dynamic) and backend context
- TypeInstance upload would ignore TypeInstance value if a given Storage Backend (we have its ID) has `acceptValue` to false [0.5MD]
  - some additional logging would be needed
- if such artifact is created manually, Content Developer should fill the `value` property based on additionalContext. This would be the runner's responsibility


Later, if we will need that, we can introduce the following container, which would enrich a given TypeInstance

```yaml
- - name: resolve-dynamic-ti
    template: resolve-dynamic-ti
    arguments:
      - name: input-artifact
        from: "{{steps.resolve-dynamic-ti.outputs.artifacts.postgresql}}"
      - name: backend
        # fortunately it is already injected into workflow
        from: "{{workflow.outputs.artifacts.helm-template-storage}}"
```

That container would fill the `value` property based on a given storage backend. However, the storage backend would need to support fetching value based purely on context, without TypeInstance ID.

It could be an additional method:

```proto
message PreCreateValueRequest {
  // no TypeInstance ID and resourceVersion
  bytes context = 1;
}

message PreCreateValueResponse {
  optional bytes value = 1;
}


service StorageBackend {
  rpc PreCreateValue(PreCreateValueRequest) returns (PreCreateValueResponse);
}
```

or, modified the existing GetValue:

```proto
message GetValueRequest {
  // optional - not needed by storage backends which don't accept value and base just on the context
  optional string type_instance_id = 1;
  optional uint32 resource_version = 2;

  bytes context = 3;
}

service StorageBackend {
  rpc GetValue(GetValueRequest) returns (GetValueResponse);
}
```









### Agreement
- Split gRPC API:
  -  Dynamic Storage
    - Remove lock/unlock
    - Add new gRPC method:

    ```proto
    message GetPreCreateValueRequest {
      // no TypeInstance ID and resourceVersion
      bytes context = 1;
    }

    message GetPreCreateValueResponse {
      optional bytes value = 1;
    }

    service StorageBackend {
      rpc GetPreCreateValue(GetPreCreateValueRequest) returns (GetPreCreateValueResponse);
    }
    ```
  - Static Storage Backend
    - Without `GetPreCreateValue` method 

- Add generic container to enrich TypeInstance (run `GetPreCreateValue`) at the end of a given workflow
  - Manually added by Content Developer
  - Add the following steps after Helm installation:

  ```yaml
  - - name: resolve-dynamic-ti-value
      template: resolve-dynamic-ti-value
      arguments:
      - name: input-artifact
        from: "{{steps.resolve-dynamic-ti.outputs.artifacts.postgresql}}"
      - name: backend
        # fortunately it is already injected into workflow
        from: "{{workflow.outputs.artifacts.helm-template-storage}}"
  ```

- Modify TypeInstance uploader: TypeInstance upload would ignore TypeInstance value if a given Storage Backend (we have its ID) has `acceptValue` to false [0.5MD]
  - some additional logging would be needed




### Dedicated property in artifact

This is a modification of the suggested solution.

Save rendered value under different property in the artifact (e.g. `.resolvedValue`) and use that in umbrella workflow

- this property `.resolvedValue` would always be saved as a part of runner
    - runners would save this property
        - Terraform runner - no need to call external backend
        - Helm runner - no need to call external backend
    - it would be saved both when dynamic and static TI is outputted, for consistency
- jinja2 and merger would need to be reworked to support unpacking `resolvedValue` instead of `value`

- if such artifact is created manually, Content Developer should fill this field, and it would be redundant with static `value`.
    - but we need to distinguish static value and generated value
- for other runners there should be a dedicated container which connects to storage backend and resolves the value:

    ```yaml
    - - name: resolve-dynamic-ti
        template: resolve-dynamic-ti
        arguments:
          - name: input-artifact
            from: "{{steps.resolve-dynamic-ti.outputs.artifacts.postgresql}}"
          - name: backend
            # fortunately it is already injected into workflow
            from: "{{workflow.outputs.artifacts.helm-template-storage}}"
    ```

- Change gRPC schema for storage backends

  ```proto
  message GetValueRequest {
    optional string type_instance_id = 1; // optional, when a given  storage backend should support this
    uint32 resource_version = 2;
    bytes context = 3;
  }

  service StorageBackend {
    rpc GetValue(GetValueRequest) returns (GetValueResponse);
  }
  ```

Summary: this could be treated as a workaround.
Level of effort: medium

Rejected because of redundancy 

## Other ideas, not considered

### Investigated in other issue

1. Change the way how we run nested workflow: maybe the Engine shouldn't render one huge workflow, but schedule and run separate Argo Workflows instead? Every workflow would have dedicated upload/download steps
    - Investigated as a part of #546

### Rejected ideas

1. Hack: Add "fetcher" container which fetches Helm template storage backend and renders the template
    - We would need to fake TypeInstance ID, because there's no TypeInstance yet
    - It won't work for other backends, which base on actual TypeInstance existing in Hub



