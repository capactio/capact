# Workflow rendering PoC

This proof of concept show a way to build a single Argo workflow from a OCF Implementation, which offloads work to other Implementations.

```
.
├── implementations # stores OCH Implementations used in this PoC
└── render          # Golang rendering source code
```

## How does it work

It uses Argo Workflows feature to handle nested workflows. A single Implementation defines an Argo workflow. If at some point, we want to offload some work to other workflow (let's call it 'imported workflow'), we can in Argo:

1. Include all templates from the imported workflow in our main workflow
2. Add a step, which references the entrypoint template of the imported workflow

To allow us reference, which workflow (which is defined in OCH Implementation) should be imported, we have to extend the syntax of the Argo workflow.

This PoC extends the Argo workflow syntax by adding a optional `action` field in the Argo Workflow Step definition:
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
      # imported argo template definitions
```

## Usage

The default SA in the k8s namespace needs proper RBAC permissions, so the Argo workflow pod can monitor other pods. You can add temporary admin access to the default SA on the default namespace with:
```bash
kubectl create clusterrolebinding default-default-admin --clusterrole admin --serviceaccount default:default
```

Render and execute PostgreSQL installation:

```bash
go run docs/investigation/workflow-rendering/main.go 'cap.implementation.bitnami.postgresql.install' | kubectl apply -f -

Render and execute Jira installation:

go run docs/investigation/workflow-rendering/main.go 'cap.implementation.atlassian.jira.install' | kubectl apply -f -
```
