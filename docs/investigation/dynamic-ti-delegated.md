# Dynamic TypeInstances in delegated storage


## Ideas


1. Save rendered value under different property in the artifact (e.g. `.resolvedValue`) and use that in umbrella workflow
    - Ignore it during TypeInstance upload
    - It probably won't work for other backends, which base on actual TypeInstance existing in Hub


```proto
message GetValueRequest {
  optional string type_instance_id = 1; # optional, storage backend should support this
  uint32 resource_version = 2;
  bytes context = 3;
}

service StorageBackend {
  rpc GetValue(GetValueRequest) returns (GetValueResponse);
}
```

- this property `.resolvedValue` would always be saved as a part of runner
    - runners would save this property
        - Terraform runner
        - Helm runner
    - it would be saved both when dynamic and static TI is outputted, for consistency
    - jinja2 and mergerer would support this
- the property
- if such artifact is created manually, Content Developer should fill this field, and it would be redundant with static `value`.
    - but we need to distinguish static value and generated value
- for other runners there should be a dedicated container which connects to storage backend and resolves the value:
- `resolvedValue`

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


Summary: this could be treated as a workaround. It would work

## Other ideas, not considered

### Investigated in other issue

1. Change the way how we run nested workflow: maybe the Engine shouldn't render one huge workflow, but schedule and run separate Argo Workflows instead? Every workflow would have dedicated upload/download steps
    - Investigated as a part of #546

### Rejected ideas

1. Hack: Add "fetcher" container which fetches Helm template storage backend and renders the template
    - We would need to fake TypeInstance ID, because there's no TypeInstance yet
    - It won't work for other backends, which base on actual TypeInstance existing in Hub
