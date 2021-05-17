# Kind configuration files

The following directory contains configuration files for kind clusters, such as Helm chart overrides.

### Regenerate self-signed CA root key and certificate

For development and integration tests kind cluster, cert-manager works as a certification authority and issues certificates for Ingress resources.

To regenerate the CA root key and certificate, run the following command:

```bash
cd ./hack/cluster-config/kind
openssl genrsa -out capact-local-ca.key 2048
openssl req -x509 -sha256 -new -nodes -key capact-local-ca.key -days 3650 -out capact-local-ca.crt
```

Once the new certificate is regenerated, update the values in [`ca-config.yaml`](./ca-config.yaml):

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ca-key-pair
data:
  tls.crt: # paste the base64 encoded certificate from capact-local-ca.crt
  tls.key: # paste the base64 encoded private key from capact-local-ca.key
```

You might need to import the newly generated CA certificate in your system's and browser's trust stores.
