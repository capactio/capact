# Kind configuration files

The following directory contains configuration files for kind clusters, such as Helm chart overrides.

### Regenerate self-signed SSL certificate

For development and integration tests kind cluster, a self-signed wildcard SSL certificate is used. The certificate is valid for all `voltron.local` subdomains.    

To regenerate it, run the following command:

```bash
cd ./hack/cluster-config/kind
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
    -keyout voltron.local.key -out voltron.local.crt \
    -subj "/CN=*.voltron.local/O=Voltron" -reqexts SAN \
    -extensions SAN -config <(cat /etc/ssl/openssl.cnf <(printf "\n[ req ]\nx509_extensions = v3_ca\n[SAN]\nsubjectAltName=DNS:voltron.local,DNS:*.voltron.local")) 
```

Once the new certificate is regenerated, update the values in [`overrides.ingress-nginx.yaml`](./overrides.ingress-nginx.yaml):

```yaml
# (..)
tlsCertificate:
    # (..)
    crt: |-
      # Paste the content of `voltron.local.crt` file
    key: |-
      # Paste the content of `voltron.local.key`
```
