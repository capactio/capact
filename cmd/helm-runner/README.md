# Helm runner

## Overview

Helm runner is a runner [Voltron runner](../../docs/runner.md), which creates and manages Helm releases.

## Prerequisites

- Running Kubernetes cluster
- Go compiler 1.14+

## Usage

Normally the runner is started by Voltron Engine, but you can run the runner locally without the Engine.

To start the runner type:
```bash
RUNNER_INPUT_PATH=cmd/helm-runner/example-input.yml \
  go run cmd/helm-runner/main.go
```

## Configuration

The following environment variables can be set:

| Name                   | Default | Description                        |
|------------------------|---------|------------------------------------|
| RUNNER_INPUT_PATH      |         | Path of the runner YAML input file |
| RUNNER_LOGGER_DEV_MODE | `false` | Enable additional log messages     |
