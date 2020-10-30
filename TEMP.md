# To remove

```bash
make dev-cluster
helm install ingress-nginx ingress-nginx/ingress-nginx --create-namespace --namespace=ingress-nginx --wait -f ./hack/ingress-kind-values.yaml 
```
``

1. Apply Pods with Services and Ingress 

    ```bash
    kubectl apply -f https://kind.sigs.k8s.io/examples/ingress/usage.yaml
    ```
1. Test whether it works:

    ```bash
    # should output "foo"
    curl localhost/foo
    # should output "bar"
    curl localhost/bar
   ```
