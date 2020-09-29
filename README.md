# Voltron

## Overview

This repository contains the Go codebase for the Voltron project.

### Project structure

The repository has the following structure:

```
  ├── cmd
  │ ├── gateway                 # GraphQL Gateway that consolidates all Voltron GraphQL APIs in one endpoint.
  │ ├── k8s-engine              # Kubernetes Voltron engine
  │ └── och                     # OCH server
  │
  ├── docs                      # Documentation related to the project
  │
  ├── hack                      # Scripts used by the Voltron developers
  │
  ├── pkg                       # Component related logic.
  │ ├── db-populator            # Populates Voltron entities to graph database.
  │ ├── engine                  # Voltron platform agnostic engine.
  │ ├── gateway                 # GraphQL Gateway
  │ ├── och                     # Open Capability Hub server 
  │ ├── runner                  # Voltron runners, e.g. Argo Workflow runner, Helm runner etc.
  │ └── sdk                     # SDK for Voltron eco-system.
  │
  │── test                      # Cross-functional test suites
  │
  └── go.mod                    # Manages Go dependency. There is a single dependency management across all components in this monorepo.
```
