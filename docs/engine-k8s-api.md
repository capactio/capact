# Engine Kubernetes API

The following document lists all decisions regarding Engine Kubernetes API.

## CRDs

The following sections lists all agreements regarding Custom Resource Definitions.

### Scope

Initially we go with only namespace-scoped Action resource. In the future, we may introduce cluster-wide ClusterAction.

### User-provided data separation

User-provided data reside in `Action.spec`, and controller-provided data in `Action.status` subresource. For example, if user wants to override the rendered Action, he/she has to copy it from `.status.renderedAction`, modify and put it in `.spec.renderedActionOverride`. 

### Status

The most common approach to represent state of the resource is to use conditions array as per [API conventions document](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status). The same document clarifies that simpler `phase` enum approach, visible e.g. in Pod status, is going to be deprecated. The document points to an issue with discussion from 2015. 
On the other hand, [`cluster-api`](https://github.com/kubernetes-sigs/cluster-api) uses `phase` approach, and it doesn't use Conditions at all, because ["they are soon to be deprecated"](https://github.com/kubernetes-sigs/cluster-api/blob/112951ee/docs/proposals/20181121-machine-api.md#conditions) (2018). Argo Workflows uses [both `conditions` and `phase` approaches](https://github.com/argoproj/argo/blob/3156559b40afe4248a3fd124a9611992e7459930/pkg/apis/workflow/v1alpha1/workflow_types.go#L1111).

To sum it up, it looks like there are different opinions how to represent the state. For more details, read the article ["What the heck are Conditions in Kubernetes controllers?"](https://dev.to/maelvls/what-the-heck-are-kubernetes-conditions-for-4je7).

Analysing our case, we found out that:
- we need to show on UI a simple high level status for a given Action. Calculating it from conditions array would be complex.
- Currently, UI and `kubectl` are the only consumers of the Action status. Conditions array wouldn't bring many benefits to us at a current state of the project.

We decided that we initially go with the `phase` approach. In the future, we may introduce conditions array, following [Kubernetes API conventions and guidelines](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status).
