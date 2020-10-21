# Engine Kubernetes API

> **NOTE**: This document is incomplete. Please come back later.
<!-- TODO: Consider conditions in status? Or describe why we didn't go with that approach right now -->

The following document lists all decisions regarding Action Custom Resource Definition.

- Currently, there is only namespace-scoped Action resource. In the future, we may introduce cluster-wide ClusterAction.
- Following best practices, user provided fields are stored in `Action.spec`, and controller-provided data in `Action.status`.
