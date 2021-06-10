# Capact

**Capact** (pronounced: "cape-act", /ˈkeɪp.ækt/) is a simple way to manage applications and infrastructure.

## Documentation

The Capact documentation can be found on the [capact.io/docs](https://capact.io/docs) website.

## Get started

The section contains useful links for getting started with Capact.

- **Introduction:** To learn what is Capact, read the [Introduction](https://staging.capact.io/docs/introduction) document.
- **Installation:** To learn how to install Capact, follow the [Installation](https://staging.capact.io/docs/installation/local) documents.
- **Development:** To run Capact on your local machine and start contributing to Capact, read the [Development](https://staging.capact.io/docs/development/development-guide) documents.

## Project structure

The repository has the following structure:

```
  .
  ├── cmd                     # Main application directory
  │
  ├── deploy                  # Deployment configurations and templates
  │
  ├── docs                    # Documents that are not published on the Capact website, such as proposals and investigations
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
