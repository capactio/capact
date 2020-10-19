# Voltron

## Overview

This repository contains the Go codebase for the Voltron project.

## Project structure

The repository has the following structure:

```
  ├── cmd
  │ ├── gateway                 # GraphQL Gateway that consolidates all Voltron GraphQL APIs in one endpoint
  │ ├── k8s-engine              # Kubernetes Voltron engine
  │ └── och                     # OCH server
  │
  ├── deploy                    # Deployment configurations and templates
  │ └── kubernetes              # Kubernetes related deployment (Helm charts, CRDs etc.)
  │
  ├── docs                      # Documentation related to the project
  │ └── investigation           # Investigations and proof of concepts files
  │
  ├── hack                      # Scripts used by the Voltron developers
  │
  ├── ocf-spec                  # Open Capability Format Specification
  │
  ├── pkg                       # Component related logic
  │ ├── db-populator            # Populates Voltron entities to graph database
  │ ├── engine                  # Voltron engine
  │ │ ├── api                   # Engine platform-agnostic api 
  │ │ └── k8s                   # Code related to k8s platform engine implementation 
  │ ├── gateway                 # GraphQL Gateway
  │ ├── och                     # Open Capability Hub server 
  │ ├── runner                  # Voltron runners, e.g. Argo Workflow runner, Helm runner etc.
  │ └── sdk                     # SDK for Voltron eco-system
  │
  │── test                      # Cross-functional test suites
  │
  ├── Dockerfile                # Dockerfile template to build applications and tests images
  │
  └── go.mod                    # Manages Go dependency. There is single dependency management across all components in this monorepo
```

## Development

Read [this](./docs/development.md) document to learn how to develop the project. 
