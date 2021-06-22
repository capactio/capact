# Capact

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GitHub Actions main branch CI](https://github.com/capactio/capact/actions/workflows/branch-build.yaml/badge.svg?branch=main)](https://github.com/capactio/capact/actions/workflows/branch-build.yaml?query=branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/capactio/capact)](https://goreportcard.com/report/github.com/capactio/capact)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/4988/badge)](https://bestpractices.coreinfrastructure.org/projects/4988)

**Capact** (pronounced: "cape-act", /ˈkeɪp.ækt/) is a simple way to manage applications and infrastructure.

> **⚠️ WARNING**: Capact versions prior to 0.4.0 are considered experimental. Capact 0.4.0, the very first public Capact release, is coming very soon, along with the **capact.io** website going live. Until this moment, the links may not work properly. Stay tuned for the announcement!

## Documentation

The Capact documentation can be found on the [capact.io](https://capact.io) website.

The documentation sources reside on the [`website`](https://github.com/capactio/website) repository under [`docs`](https://github.com/capactio/website/tree/main/docs) directory.

## Get started

The section contains useful links for getting started with Capact.

- **Introduction:** To learn what is Capact, read the [Introduction](https://capact.io/docs/introduction) document.
- **Installation:** To learn how to install Capact, follow the [Installation](https://capact.io/docs/installation/local) documents.
- **Development:** To run Capact on your local machine and start contributing to Capact, read the [Development](https://capact.io/docs/development/development-guide) documents.

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
  ├── hub-js                  # Node.js implementation of Capact Hub
  │
  ├── pkg                     # Public component and SDK code
  │
  ├── test                    # Cross-functional test suites
  │
  ├── Dockerfile              # Dockerfile template to build applications and tests images
  │
  └── go.mod                  # Manages Go dependency. There is single dependency management across all components in this monorepo
```
