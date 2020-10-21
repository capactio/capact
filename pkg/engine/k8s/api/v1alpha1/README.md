<!-- TODO: Move this document -->

# Action CRD

The following document lists all decisions regarding Action Custom Resource Definition.

- Currently there is only namespace-scoped Action resource. In future, we may introduce cluster-wide ClusterAction.
- Following best practices, user provided fields are stored in `Action.spec`, and controller-provided data in `Action.status`.
