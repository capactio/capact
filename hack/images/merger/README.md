# merger

## Overview

This folder contains the Docker image which merges multiple input YAML files into a single one.

The Docker image contains the `merger.sh` helper script. The script is an entrypoint of the image, and it is used to prefix and merge all YAML files found in `$SRC` directory. Additionally, if the YAML file contains the `value` key, then it is unpacked from that key. Each file is prefixed with a file name without extension.

## Installation

To build the Docker image, run this command:

```bash
docker build -t merger .
```

## Configuration

You can configure the merger script by passing the following environment variables:

| Variable                  | Default      | Description                                      |
| ------------------------- | ------------ | ------------------------------------------------ |
| SRC                       | /yamls       | Path to the directory with YAML files.           |
| OUT                       | /merged.yaml | Output file with prefixed and merged YAML files. |
