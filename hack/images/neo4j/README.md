# neo4j

This folder contains Dockerfile which wraps official [Neo4j Docker image](https://hub.docker.com/_/neo4j/) and adds `apoc` plugin. As a result, when Neo4j container is started, plugin is not downloaded from the internet which allows air-gaped usage.

The [`neo4jlabs-plugins.json`](./neo4jlabs-plugins.json) and [`docker-entrypoint.sh`](./docker-entrypoint.sh) were copied from [this](https://github.com/neo4j/docker-neo4j/pull/302) PR and adjusted for `apoc` plugin. We had to copy them as the 4.4 version was not yet release, but we want to re-use the option to load plugins from disk instead of downloading them from the internet each time.

To update our Neo4j image, run:
```bash
docker build -t ghcr.io/capactio/neo4j:4.2.8-apoc .
docker push ghcr.io/capactio/neo4j:4.2.8-apoc
```

> **NOTE:** You need to be logged to ghcr.io.
