# Helm runner

## Overview

Helm runner is a runner [Voltron runner](../../docs/runner.md), which creates and manages Helm releases.

## Prerequisites

- Running Kubernetes cluster
- Go compiler 1.14+
- Helm 3+

## Usage

Normally the runner is started by Voltron Engine, but you can run the runner locally without the Engine.

To start the runner type:
```bash
RUNNER_INPUT_PATH=cmd/helm-runner/example-input.yml \
  go run cmd/helm-runner/main.go
```

Running this should create a Helm PostgreSQL release:
```bash
$ helm list
NAME                    NAMESPACE       REVISION        UPDATED                                 STATUS       CHART                    APP VERSION
postgresql-1607608471   default         1               2020-12-10 14:54:34.882358554 +0100 CET deployed     postgresql-10.1.3        11.10.0 
```

## Configuration

The following environment variables can be set:

| Name                   | Default | Description                        |
|------------------------|---------|------------------------------------|
| RUNNER_INPUT_PATH      |         | Path to the runner YAML input file |
| RUNNER_LOGGER_DEV_MODE | `false` | Enable additional log messages     |
