# Versioning

This document describes Voltron versioning strategy.

## Table of contents

<!-- toc -->

- [Overview](#overview)
- [Diagram](#diagram)
- [OCF Version](#ocf-version)
  * [Supporting multiple OCF versions in OCH](#supporting-multiple-ocf-versions-in-och)
  * [Deprecation policy of OCF versions](#deprecation-policy-of-ocf-versions)
- [Manifests revision](#manifests-revision)
- [TypeInstance `resourceVersion`](#typeinstance-resourceversion)
- [Core manifests](#core-manifests)
- [The `appVersion`](#the-appversion)
  * [Default versions for appVersion in SemVer format](#default-versions-for-appversion-in-semver-format)
  * [Default versions for appVersion in different format than SemVer](#default-versions-for-appversion-in-different-format-than-semver)
  * [Conflict prevention](#conflict-prevention)
- [Engine and CLI versions](#engine-and-cli-versions)
- [SDK Version](#sdk-version)

<!-- tocstop -->

## Overview

The versioning for OCF and OCH are similar in concept to how Kubernetes implements versioning. Below is a table
comparing Voltron versioning to Kubernetes versioning.

| Voltron                        | Kubernetes            |
| ------------------------------ | --------------------- |
| OCH Version                    | Kubernetes Version    |
| OCF Version                    | Resource `apiVersion` |
| Manifests `revision`           | `resourceVersion`     |
| TypeInstance `resourceVersion` | `resourceVersion`     |
| Engine/CLI                     | `kubectl`             |
| Go SDK                         | `client-go`           |

## Diagram

The following diagram shows the versioning concept:

![Versioning concept](./assets/versioning.svg)

## OCF Version

This is the version of the Open Capability Format itself. The version changes every time there is a change in any of OCF
entity manifest, such as adding or removing properties in Implementation manifest or introducing a brand-new entity.

### Supporting multiple OCF versions in OCH

OCH supports multiple versions of OCF. To achieve that, we reuse the API versioning concept from Kubernetes. A single
OCF version is used to store resources in the database. However, OCH does the conversion between the stored
resource and one of the supported OCF versions by OCH on-the-fly.

The cluster administrator migrates the storage version of the resource manually during OCH upgrade/downgrade. In the future,
we will introduce automatic migration between storage versions or external tools that facilitates the process.

**Example:**

OCH 0.3.0 can support OCF versions 0.2.0 and 0.1.0. The RepoMetadata entity is defined as follows:

```yaml
kind: RepoMetadata
# (...)
spec:
   ochVersion: 0.1.0
   ocfVersion:
       default: 0.2.0
       supported:
           - 0.1.0
           - 0.2.0
# (...)
```

The manifest version stored in OCH is 0.2.0. However, using a different API endpoint, the user can fetch manifests in
version 0.1.0. OCH supports on-the-fly conversion between the default (stored) OCF manifests to the OCF manifests in
supported versions.

### Deprecation policy of OCF versions

The deprecation policy is very similar to
the [Kubernetes deprecation policy](https://kubernetes.io/docs/reference/using-api/deprecation-policy/). The only change
is that we will use [Semantic Versioning](https://semver.org/) for versioning OCF. There are multiple reasons to use SemVer 2 for OCF versioning:

- Unification of versioning across all Project Voltron components.
- Clear way to represent new features without breaking changes. You can easily see that there is a new OCF feature that
  you can use. In Kubernetes API versioning, a new non-breaking feature doesn’t change the version, e.g. `v1`.
- Unification with other projects in the open source community, such as OAM, CNAB, CloudEvents, AsyncAPI.

Once we deprecate an OCF version, we will include deprecation notices in OCH release notes. We will warn users every
time they access deprecated OCF:

- in CLI,
- on UI,
- through the GraphQL API (using [Warning header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Warning)).

A deprecated OCF version will have a transition period (
see [Kubernetes deprecation policy](https://kubernetes.io/docs/reference/using-api/deprecation-policy/) for details) where it’s
still supported by OCH. After that period, we remove support for the deprecated OCF version from the OCH release.

OCH rejects any submission of OCF manifests in unsupported versions.

The OCF storage version in OCH is never the deprecated one. This way, there won’t be a case when existing manifests are
removed during OCH upgrade due to an unsupported version.

## Manifests revision

Unlike Kubernetes `resourceVersion`, we store all previous revisions of manifests such Type, Implementation, Interface and Attribute. Users can consume them anytime. Manifests can refer to older revisions of the other manifests.

**Example:** Implementation implements specific Interface revision. Revision of the Interface is increased once input or output Type changes. Content Creator updates the revision manually.

## TypeInstance `resourceVersion`

This is the version of TypeInstance metadata and spec fields. Unlike Kubernetes, we store historical data for audits and rollback purposes.

## Core manifests

We version core manifests (manifests, which are located under [`core` subdirectory](../och-content/core)) in the same way as the OCF itself. Core entities are strictly tied into the OCH release, and they are read-only.

## The `appVersion`

> **NOTE:** The following subsection describes future `appVersion` features, which are not yet implemented.

The Implementation manifest contains `appVersion` field, which defines the supported version of the actual application. The `appVersion` field is independent from the revision.

The appVersion is an object in the following format:

```
appVersion: "1.0.x, 1.1.0 - 1.3.0" # A string with allowed version ranges
```

It is inspired by
the [CNAB dependency version object](https://github.com/cnabio/cnab-spec/blob/master/500-CNAB-dependencies.md). If the
appVersion ranges are in SemVer 2 format, you can use ranges using dashes. If the `appVersion` ranges are not in SemVer 2,
then you have to specify every supported appVersion in the string.

### Default versions for appVersion in SemVer format

During the submission of the Implementation manifest, if the `appVersion` is defined in the SemVer format, OCH updates the
following versions:

- `latest` — depending on the OCH configuration, it points to stable or edge version
- `stable` — points to the Implementation with highest semVer version in range without
  suffix [starting from the hyphen](https://semver.org/#spec-item-9)
  
  For example, if the range is defined as `1.0.x - 1.1.0-beta1`, the 1.0.9 is picked as an `appVersion`.
  
- `edge` — points to the Implementation with highest semVer version in range, even if it contains suffix starting from
  hyphen
  
  For example, if the range is defined as `1.0.x - 1.1.0-beta1`, the 1.1.0-beta1 is picked as an `appVersion`.

You can use the versions in Implementation manifest to filter prerequisite Implementations based on the appVersion
value. For example, if your Implementation depends on the latest stable PostgreSQL version, then you can use the `stable`
version as the `appVersion` of PostgreSQL.

### Default versions for appVersion in different format than SemVer

If the application is not versioned using SemVer format, we assume that all possible `appVersion` values are sorted from
oldest to newest. This way, the `latest` version is always the newest appVersion value.

**Example:**

In the following example, the `baz` version is picked as the latest one.

```yaml
ocfVersion: 0.0.1
revision: 1.0.0
metadata:
  prefix: cap.implementation.database.mysql
  name: install
spec:
   appVersion: "foo, bar, baz"
```

### Conflict prevention

An `appVersion` can be defined as a range. During Implementation manifest submission, the OCH validates whether the
`appVersion` range doesn't overlap with the same revision of the Implementation manifest. As noted earlier, the manifest
`revision` is independent from the `appVersion`.

**Example:**

1. The following implementation manifest already exists in the OCH:

```yaml
ocfVersion: 0.0.1
kind: Implementation
metadata:
  prefix: cap.implementation.database.mysql
  name: install
revision: 1.0.0
spec:
  appVersion: "8.0.0-8.0.20"
```

2. When you try to submit the following manifest to the OCH...

```yaml
ocfVersion: 0.0.1
metadata:
  prefix: cap.implementation.database.mysql
  name: install
revision: 1.0.1
spec:
  appVersion: "8.0.x"
```

3. ...the operation fails as the `appVersion` range overlaps with existing `cap.implementation.database.mysql.install` Implementation manifest.

## Engine and CLI versions

Engine and CLI versions need to compatible with OCH, as they consume content from OCH. This is similar case as the `kubectl` is compatible with the Kubernetes APIServer. 

## SDK Version

The SDK is always released with a new version of OCH. The SDK has a reference to the OCH in a given version. The SDK is supported within the most recent three minor releases of OCH.
