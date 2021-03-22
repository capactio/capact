# Terminology

This document lists and explains all terms used through the Voltron documentation.

> There are only two hard things in Computer Science: cache invalidation and naming things.
> -- Phil Karlton

## Table of contents

<!-- toc -->

- [Terminology](#terminology)
  - [Table of contents](#table-of-contents)
  - [Common terms](#common-terms)
    - [Capability](#capability)
    - [Action](#action)
    - [Runner](#runner)
  - [Components](#components)
  - [Entities](#entities)

<!-- toc-stop -->

## Common terms

### Capability

Characteristic of an application, described with an Interface or Implementation. A capability may be a prerequisite (dependency) for other Implementations.

### Action

Task that the Engine schedules, and the Runner executes. Action is usually in a form of an arbitrary workflow.

### Runner

Action, which handles execution of other Action. Runner is usually defined in form of Interface and Implementation. 

There is also a built-in Runner, which is built-in into platform-specific Engine implementation. It is defined with only abstract Interface and doesn't have Implementation manifest.

To learn more about runners, see the dedicated [`runner.md`](./runner.md) document.

## Components

There are the following components in the system:

- [OCF](./e2e-architecture.md#ocf)
- [UI](./e2e-architecture.md#ui)
- [CLI](./e2e-architecture.md#cli)
- [Gateway](./e2e-architecture.md#gateway)
- [Engine](./e2e-architecture.md#engine)
- [OCH](./e2e-architecture.md#och)
- [SDK](./e2e-architecture.md#sdk)

## Entities

There are the following entities in the system:

- [Attribute](../ocf-spec/0.0.1/README.md#attribute)
- [Implementation](../ocf-spec/0.0.1/README.md#implementation)
- [Interface](../ocf-spec/0.0.1/README.md#interface)
- [RepoMetadata](../ocf-spec/0.0.1/README.md#repo-metadata)
- [Type](../ocf-spec/0.0.1/README.md#type)
- [Vendor](../ocf-spec/0.0.1/README.md#vendor)