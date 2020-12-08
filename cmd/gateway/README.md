# Voltron GraphQL gateway

## Supported features:

- GraphQL reverse proxy for multiple GraphQL endpoints
- basic auth authentication

## How to build

```bash
make build-app-image-gateway
```

## How to use

To deploy the gateway to your dev cluster type:
```bash
make dev-cluster-update
```

This will build all apps (including gateway) and deploy them to the dev KinD cluster. It also adds a entry in `/etc/hosts`:
```properties
# /etc/hosts
[...]
127.0.0.1 gateway.voltron.local
```

You can access the GraphQL playground on the gateway by opening http://gateway.voltron.local/graphql. As currently the gateway is secured using basic auth you need to provide the following headers:
```json
{
  "Authorization": "Basic Z3JhcGhxbDp0MHBfczNjcjN0"
}
```

Then you should be able to make queries to the gateway:
```graphql
query($implementationPath: NodePath!) {
  implementation(path: $implementationPath) {
    name,
    prefix,
    latestRevision {
      spec {
        action {
          runnerInterface
          args
        }
      }
    }
  }
}
```

## Hacking

Main source code is in:
- `cmd/gateway` - binary main
- `internal/gateway` - private gateway code
- `pkg/gateway` - public gateway code
