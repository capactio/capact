
```yaml
kind: Implementation
metadata:
  displayName: "Boostrap PostgreSQL cluster"
  attributes:
    cap.attribute.containerization.kubernetes.kubeconfig-support:
      revision: 0.1.0

spec:
  requires:
    cap.core.type.platform:
      oneOf:
        - name: kubernetes
          alias: k8s
          revision: 0.1.0

  action:
    runnerInterface: argo.run
    args:
      workflow:
        entrypoint: postgresql-cluster
        templates:
          - name: postgresql-cluster
            inputs:
              artifacts:
                - name: input-parameters
                - name: k8s
            steps:
              - - name: create-target-cluster
                  capact-when: k8s.kubeconfig == nil # Add support to check fields
                  capact-action: k8s.deploy
              - - name: postgresql-install
                  capact-action: postgresql.install
                  # capact-target: "{{steps.create-target-cluster.outputs.artifacts.kubeconfig}}" # optional syntax sugar
                  capact-policy:
                    rules:
                      - interface:
                          path: postgresql.install
                        oneOf:
                          - implementationConstraints:
                              requires:
                                - path: "cap.core.type.platform.kubernetes"
                            inject:
                              requiredTypeInstances:
                                - artifact: "{{steps.create-target-cluster.outputs.artifacts.k8s}}"
```

Requires:
1. Add option to check the TypeInstance properties via `capact-when`.
2. Merge the `cap.type.containerization.kubernetes.kubeconfig` to `cap.core.type.platform.kubernets` as we don't have Type composition yet.
3. Solve issue https://github.com/capactio/capact/issues/537 but also take into account TypeInstances from the `requires` section.

Problems:
1. Which step of umbrella workflow should be executed on external cluster?
2. Not readable (almost magic) `k8s.kubeconfig == nil` check.



What about such example:
```yaml
  requires: ## alternatives, doesn't make sens if we will have, but not only engine also polices needs to be done
    # kubernetes -> sa
    # kubernetes -> kubeconfig (how to select strategy e.g. we need add `multicluster.admiralty.io/elect: ""` with nodeSelector)
    oneOf:
      - name: cap.core.type.platform.kubernetes
        revision: 0.1.0
      - name: cap.type.containerization.kubernetes.kubeconfig
        revision: 0.1.0
```
