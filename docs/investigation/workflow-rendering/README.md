# Workflow rendering PoC

This proof of concept show a way to build a single Argo workflow from a OCF Implementation, which references to other Implementations.

```
.
├── implementations # stores OCH Implementations used in this PoC
└── render          # Golang rendering source code
```

## How does it work

This proof of concept renders a final Argo workflow by merging all child Argo workflows into the root workflow. It works using the following algorithm:

1. Include all templates from a child workflow in the root workflow
2. Modify the step, where the child workflow was referenced, to reference the entrypoint template of the child workflow

To allow us reference, which workflow (which is defined in OCH Implementation) should be imported, we have to extend the syntax of the Argo workflow.

This PoC extends the Argo workflow syntax by adding a optional `action` field in the Argo Workflow step definition:
```yaml
workflow:
  entrypoint: main
  templates:
    - name: main
      steps:
        - - name: offload-work
            action:
              name: path.to.imported.implementation
```

In the rendering process, this gets changed to:
```yaml
workflow:
  entrypoint: main
  templates:
    - name: main
      steps:
        - - name: offload-work
            template: imported-wf-entrypoint

    - name: imported-wf-entrypoint
      # imported Argo template definitions
```

## Usage

The default ServiceAccount in the `default` Namespace needs proper RBAC permissions, so the Argo workflow pod can run other pods. You can add temporary admin access to the default SA on the default namespace with:
```bash
kubectl create clusterrolebinding default-default-admin --clusterrole admin --serviceaccount default:default
```

Render and execute PostgreSQL installation:

```bash
go run docs/investigation/workflow-rendering/main.go 'cap.implementation.bitnami.postgresql.install' | kubectl apply -f -
```

Render and execute Jira installation:

```bash
go run docs/investigation/workflow-rendering/main.go 'cap.implementation.atlassian.jira.install' | kubectl apply -f -
```
