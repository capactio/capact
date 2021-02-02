# OCF content creation guide

This guide shows first steps on how to develop OCF content for Voltron. We will show how to:
- define new Types and Interfaces,
- create Implementation for the Interfaces,
- use other Interfaces in your Implementations.

As an example, we will create OCF manifests to deploy BitBucket with a PostgreSQL database.

## Getting started

1. Clone the Voltron repository with the current OCF content.
```bash
git clone https://github.com/Project-Voltron/go-voltron.git
```

## Define the Interfaces and Types

As first, you need to create an **InterfaceGroup** manifest, which groups Interfaces coresponding to some application.
Let's create a InterfaceGroup called `cap.interface.productivity.bitbucket`, which will group Interfaces operating on BitBucket instances. In `och-content/interface/productivity/`, create a file called `bitbucket.yaml`, with the following content:
```yaml
ocfVersion: 0.0.1
revision: 0.1.0
kind: InterfaceGroup
metadata:
  prefix: cap.interface.productivity
  name: bitbucket
  displayName: "BitBucket"
  description: "Bitbucket gives teams one place to plan, collaborate, test, and deploy their code"
  documentationURL: https://support.atlassian.com/bitbucket-cloud/
  supportURL: https://support.atlassian.com/bitbucket-cloud/
  iconURL: https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png # TODO change this
  maintainers:
    - email: your.email@example.com
      name: your-name
      url: your-website

signature:
  och: eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9
```

After we have the InterfaceGroup, let's create the Interface, for installing BitBucket.
Create the directory `./och-content/interface/productivity/bitbucket`. Inside this directory, create a file `install.yaml` with the following content:

```yaml
ocfVersion: 0.0.1
revision: 0.1.0
kind: Interface
metadata:
  prefix: cap.interface.productivity.bitbucket
  name: install
  displayName: "Install"
  description: "Install action for Jira"
  documentationURL: https://support.atlassian.com/jira-software-cloud/resources/
  supportURL: https://www.atlassian.com/software/jira
  iconURL: https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png
  maintainers:
    - email: team-dev@projectvoltron.dev
      name: Voltron Dev Team
      url: https://projectvoltron.dev

spec:
  input:
    parameters:
      jsonSchema:
        value: |-
          {
            "$schema": "http://json-schema.org/draft-07/schema",
            "$ocfRefs": {
              "inputType": {
                "name": "cap.type.productivity.bitbucket.install-input",
                "revision": "0.1.0"
              }
            },
            "allOf": [ { "$ref": "#/$ocfRefs/inputType" } ]
          }
  output:
    typeInstances:
      jira-config:
        typeRef:
          path: cap.type.productivity.bitbucket.config
          revision: 0.1.0

signature:
  och: eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9
```
