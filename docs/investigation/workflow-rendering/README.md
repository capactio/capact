# Workflow rendering PoC

## Overview

This proof of concept shows a way to render an Argo workflow from a Capact Action CR. It supports conditional execution, based on provided TypeInstances and allows us to reference other Interfaces.

```
.
├── manifests             # OCF Manifests used in the PoC
│   ├── implementations
│   ├── interfaces
│   └── typeinstances
└── render                # Rendering algorithm source code
```

## Prerequisites

- [Go](https://golang.org)
- Kubernetes cluster with [Argo](https://argoproj.github.io/) installed

You have to add some RBAC permission on for the default service account on the namespace, where Argo runs the workflow pods. To do this, execute:
```
kubectl create rolebinding default-admin --clusterrole=admin --serviceaccount=default:default -n default
```

Currently, there is a problem in Argo with referencing global artifacts created in nested workflows ([GitHub issue](https://github.com/argoproj/argo-workflows/issues/4772)).

To make this PoC work you need to apply the following diff to the Argo v2.11.7 repository:

<details><summary>Git diff</summary>

```
diff --git a/workflow/controller/operator.go b/workflow/controller/operator.go
index 583d6fd7..0ea36620 100644
--- a/workflow/controller/operator.go
+++ b/workflow/controller/operator.go
@@ -47,7 +47,6 @@ import (
 	argosync "github.com/argoproj/argo/workflow/sync"
 	"github.com/argoproj/argo/workflow/templateresolution"
 	wfutil "github.com/argoproj/argo/workflow/util"
-	"github.com/argoproj/argo/workflow/validate"
 )

 // wfOperationCtx is the context for evaluation and operation of a single workflow
@@ -213,24 +212,24 @@ func (woc *wfOperationCtx) operate() {
 			return
 		}
 		woc.eventRecorder.Event(woc.wf, apiv1.EventTypeNormal, "WorkflowRunning", "Workflow Running")
-		validateOpts := validate.ValidateOpts{ContainerRuntimeExecutor: woc.controller.GetContainerRuntimeExecutor()}
-		wftmplGetter := templateresolution.WrapWorkflowTemplateInterface(woc.controller.wfclientset.ArgoprojV1alpha1().WorkflowTemplates(woc.wf.Namespace))
-		cwftmplGetter := templateresolution.WrapClusterWorkflowTemplateInterface(woc.controller.wfclientset.ArgoprojV1alpha1().ClusterWorkflowTemplates())
+		//validateOpts := validate.ValidateOpts{ContainerRuntimeExecutor: woc.controller.GetContainerRuntimeExecutor()}
+		//wftmplGetter := templateresolution.WrapWorkflowTemplateInterface(woc.controller.wfclientset.ArgoprojV1alpha1().WorkflowTemplates(woc.wf.Namespace))
+		//cwftmplGetter := templateresolution.WrapClusterWorkflowTemplateInterface(woc.controller.wfclientset.ArgoprojV1alpha1().ClusterWorkflowTemplates())

 		// Validate the execution wfSpec
-		wfConditions, err := validate.ValidateWorkflow(wftmplGetter, cwftmplGetter, woc.wf, validateOpts)
-
-		if err != nil {
-			msg := fmt.Sprintf("invalid spec: %s", err.Error())
-			woc.markWorkflowFailed(msg)
-			woc.eventRecorder.Event(woc.wf, apiv1.EventTypeWarning, "WorkflowFailed", msg)
-			return
-		}
-		// If we received conditions during validation (such as SpecWarnings), add them to the Workflow object
-		if len(*wfConditions) > 0 {
-			woc.wf.Status.Conditions.JoinConditions(wfConditions)
-			woc.updated = true
-		}
+		//wfConditions, err := validate.ValidateWorkflow(wftmplGetter, cwftmplGetter, woc.wf, validateOpts)
+
+		//if err != nil {
+		//	msg := fmt.Sprintf("invalid spec: %s", err.Error())
+		//	woc.markWorkflowFailed(msg)
+		//	woc.eventRecorder.Event(woc.wf, apiv1.EventTypeWarning, "WorkflowFailed", msg)
+		//	return
+		//}
+		//// If we received conditions during validation (such as SpecWarnings), add them to the Workflow object
+		//if len(*wfConditions) > 0 {
+		//	woc.wf.Status.Conditions.JoinConditions(wfConditions)
+		//	woc.updated = true
+		//}

 		woc.workflowDeadline = woc.getWorkflowDeadline()
```

</details>

## Usage

You can generate the workflows in this proof of concept:
1. PostgreSQL install using Helm
2. JIRA install with Helm and a provided PostgreSQL TypeInstance
3. JIRA install and a nested PostgreSQL install wit Helm

### PostgreSQL installation

To generate and run the workflow, execute:
```bash
go run main.go -och-dir=../../../och-content/ -render-input=inputs/1-postgres.yml | kubectl apply -n default -f -
```

### JIRA installation with a provided PostgreSQL TypeInstance

We will use the PostgreSQL installation created in [PostgreSQL install using Helm](#postgresql-installation), so if you did not take that step, do it now.

Update the `.spec.value.host` in `manifests/typeinstances/postgresql-1.yaml` with the PostgreSQL service Kubernetes DNS, to reflect the connection details to the PostgreSQL server:
```yaml
spec:
  value:
    superuser:
      username: "postgres"
      password: "s3cr3t"
    defaultDBName: "test"
    host: "{ postgresql-service-name }"
    port: 5432
```

To generate and run the workflow, execute:
```bash
go run main.go -och-dir=../../../och-content/ -render-input=inputs/2-jira.yml | kubectl apply -n default -f -
```

### JIRA installation with embedded PostgreSQL

Remove the JIRA Helm release from [JIRA install with Helm and a provided PostgreSQL TypeInstance](#jira-installation-with-a-provided-postgresql-typeinstance), if you did that step before.

To generate and run the workflow, execute:
```bash
go run main.go -och-dir=../../../och-content/ -render-input=inputs/3-jira-with-postgres.yml | kubectl apply -n default -f -
```

### JIRA installation with CloudSQL

Update the `.spec.value` property in `manifests/typeinstances/gcp-sa-4.yaml` with the [GCP service account key](https://cloud.google.com/iam/docs/creating-managing-service-account-keys). The service account key must have the `Cloud SQL Admin` role.

To generate and run the workflow, execute:
```bash
go run main.go -och-dir=../../../och-content/ -render-input=inputs/3-jira-with-cloudsql.yml | kubectl apply -n default -f -
```

Navigate to your [GCP CloudSQL Console](https://console.cloud.google.com/sql/instances). You should see that a new CloudSQL instance is being created.
 
## Rendering algorithm

Input:
- `manifestReference` <- Manifest reference to Interface
- `inputParameters` <- Input parameters to the Action
- `inputTypeInstances` <- List of Type Instances

1. `workflow`, _ <- `render("", manifestReference)`
2. Set `inputParameters` as Argo artifact argument for the workflow.
3. Foreach `inputTypeInstance` in `inputTypeInstances`:
   1. Add `inputTypeInstance` fetch steps at the start of the entrypoint `workflow` template.
4. Repeat:
   1. Evaluate and remove conditional steps in `workflow`.
   2. Check, if there are remaining steps with `action` key. If no, then finish rendering.
   3. Foreach `template` in `workflow`:
      1. `artifactRenameMap` <- Initialize global artifacts rename map.
      2. Foreach `step` in `template`:
         1. If `step` has `action` key:
            1. `nestedWorkflow`, `artifactRenames` <- `render(template.name-step.name, action)`
            2. Add all renames from `artifactRenames` to `artifactRenameMap`.
            3. Add all templates from `nestedWorkflow` to `workflow`
            4. Remove `action` key from `step`
         2. Foreach `artifact` in `step.arguments.artifacts`:
            1. If `artifact.from` in `artifactRenameMap`:
               1. Replace `artifact.from` based on `artifactRenameMap`


The `render(prefix, manifestReference)` function is defined following:

1. Get an `implementation` based on the `manifestReference`.
2. Create a `workflow` from the `implementation`.
3. `artifactsRenameMap` <- Initialize a global artifacts rename map.
4. Foreach `template` in `workflow`:
   1. If `prefix` is not empty:
      1. Prefix the `template.name` and all global output artifacts with `prefix`. Save artifacts rename mapping in `artifactsRenameMap`
   2. Foreach `typeInstanceOutput` in `template.TypeInstanceOutput`:
      1. Append a output type instance template to the `workflow`.
      2. Add a step to run the output type instance template at end of `template`
   3. Unset the `template.TypeInstanceOutput` field
5. If `prefix` is not empty:
   1. Prefix the `workflow.Entrypoint` with `prefix`
6. Return `workflow, artifactsRenameMap`
