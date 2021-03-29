#  Argo Workflows Chart

We need to use our own image for `argoproj/workflow-controller` until [this](https://github.com/argoproj/argo/issues/4772) bug is fixed. The [Argo Workflows Chart](https://github.com/argoproj/argo-helm/tree/cf399e6ddaa3cdbfae5c0bd454bd3cfe040f2998/charts/argo) doesn't support overriding just the `argoproj/workflow-controller` image. As a result, we had to mirror also other images to our GCR.
 
To replace the Argo images, we followed these steps: 

1. Mirror `argoproj/argocli`.

    ```bash
    docker pull argoproj/argocli:v2.12.10
    docker tag argoproj/argocli:v2.12.10 gcr.io/projectvoltron/argoproj/argocli:v2.12.10
    docker push gcr.io/projectvoltron/argoproj/argocli:v2.12.10
    ```

1. Mirror `argoproj/argoexec`.

    ```bash
    docker pull argoproj/argoexec:v2.12.10
    docker tag argoproj/argoexec:v2.12.10 gcr.io/projectvoltron/argoproj/argoexec:v2.12.10
    docker push gcr.io/projectvoltron/argoproj/argoexec:v2.12.10
    ```

1. Build and push `argoproj/workflow-controller` based on our fork. More info can be found [here](https://github.com/Project-Voltron/argo-workflows/pull/1).
