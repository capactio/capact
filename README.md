# Capact

**Capact** (pronounced: "cape-act", /ˈkeɪp.ækt/) is a simple way to manage applications and infrastructure.

## Get started

The section contains useful links for getting started with Capact.

- **Introduction:** To learn what is Capact, read the [Introduction](./docs/introduction.md) document.
- **Installation:** To learn how to install Capact, follow the [installation](./docs/installation) documents.
- **Development:** To run Capact on your local machine and start contributing to Capact, read the [`development`](./docs/development) documents.

To read the full documentation, navigate to the [capact.io/docs](https://capact.io/docs) website.

## Project structure

The repository has the following structure:

```
  .
  ├── cmd                     # Main application directory
  │
  ├── deploy                  # Deployment configurations and templates
  │
  ├── docs                    # Documentation related to the project
  │
  ├── hack                    # Scripts used by the Capact developers
  │
  ├── internal                # Private component code
  │
  ├── ocf-spec                # Open Capability Format Specification
  │
  ├── och-content             # OCF Manifests for the Open Capability Hub
  │
  ├── och-js                  # Node.js implementation of Open Capability Hub
  │
  ├── pkg                     # Public component and SDK code
  │
  ├── test                    # Cross-functional test suites
  │
  ├── Dockerfile              # Dockerfile template to build applications and tests images
  │
  └── go.mod                  # Manages Go dependency. There is single dependency management across all components in this monorepo
```
