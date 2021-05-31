#  Argo Workflows Chart

We need to use our own image for `argoproj/workflow-controller` until [this](https://github.com/argoproj/argo/issues/4772) bug is fixed. The [Argo Workflows Chart](https://github.com/argoproj/argo-helm/tree/cf399e6ddaa3cdbfae5c0bd454bd3cfe040f2998/charts/argo) doesn't support overriding just the `argoproj/workflow-controller` image. As a result, we had to mirror also other images to our GCR.
 
To replace the Argo images, we followed these steps: 

1. Mirror `argoproj/argocli`.

    ```bash
    docker pull argoproj/argocli:v2.12.10
    docker tag argoproj/argocli:v2.12.10 ghcr.io/capactio/argoproj/argocli:v2.12.10
    docker push ghcr.io/capactio/argoproj/argocli:v2.12.10
    ```

1. Mirror `argoproj/argoexec`.

    ```bash
    docker pull argoproj/argoexec:v2.12.10
    docker tag argoproj/argoexec:v2.12.10 ghcr.io/capactio/argoproj/argoexec:v2.12.10
    docker push ghcr.io/capactio/argoproj/argoexec:v2.12.10
    ```

1. Build and push `argoproj/workflow-controller` based on our fork.

    1. Clone forked version of Argo workflows and checkout to `disable-global-artifacts-validation`:
        ```bash
        git clone git@github.com:Project-Voltron/argo-workflows.git
        cd argo-workflows
        git checkout disable-global-artifacts-validation
        ```

    1. Build image:
        ```
        make controller-image
        ```
        > **NOTE:** If you will have problem with permission check this comment: https://github.com/golang/go/issues/14213#issuecomment-229815144
    
    1. Tag image:
        ```
        docker tag argoproj/workflow-controller:latest ghcr.io/capactio/argoproj/workflow-controller:v2.12.10-disable-global-artifacts-validation
        ```
    
    1. Push image
        ```
        docker push ghcr.io/capactio/argoproj/workflow-controller:v2.12.10-disable-global-artifacts-validation
        ```
    
    > **NOTE:** More info can be found [here](https://github.com/Project-Voltron/argo-workflows/pull/1).
