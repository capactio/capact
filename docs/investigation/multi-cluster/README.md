# Multi-cluster support


## Scenarios

### Use cluster generated in umbrella workflow 


1. Remove `cap.attribute.containerization.kubernetes.kubeconfig-support` from **attributes**
2. Merge the `cap.type.containerization.kubernetes.kubeconfig` to `cap.core.type.platform.kubernets` as we don't have Type composition yet.
3. Use policies to inject other `cap.core.type.platform.kubernets`
4. Add option to use `inject.requiredTypeInstances` via artifact name

```yaml
rules: # Configures the following behavior for Engine during rendering Action
  - interface: # Rules for Interface with exact path in exact revision
      path: "cap.interface.database.postgresql.install"
      revision: "0.1.0"
    oneOf: # Engine follows the order of the Implementation selection,
      # finishing when at least one matching Implementation is found
      - implementationConstraints: # In first place, find and use an Implementation which:
          requires: # AND has the following Type references defined in the `spec.requires` property:
            - path: "cap.core.type.platform.kubernetes"
              # in any revision
        inject:
          requiredTypeInstances: # For such Implementation, inject the following TypeInstances if matching Type Reference is used in `Implementation.spec.requires` property along with `alias`: 
            # Find Type Reference for the given TypeInstance ID. Then, find the alias of the Type reference in `spec.requires` property.
            # If it is defined, inject the TypeInstance with ID `9038dcdc-e959-41c4-a690-d8ebf929ac0c` under this alias.
            - id: 9038dcdc-e959-41c4-a690-d8ebf929ac0c
              description: "BigBang EU cluster" # optional
```

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
                  optional: true
            steps:
              - - name: create-target-cluster
                  capact-when: k8s == nil
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
                                - artifact: "{{steps.create-target-cluster.outputs.artifacts.kubeconfig}}"
```


Action to prepare that cluster for multi-support .. 

Injecting:
```yaml
additionalInput:
  typeInstances:
    kubeconfig:
      typeRef:
        path: cap.type.containerization.kubernetes.kubeconfig
        revision: 0.1.0
      verbs: [ "get" ]
```


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

### Pass already registered cluster

```yaml
kind: Implementation
metadata:
  displayName: Install PostgreSQL database
  attributes:
    cap.attribute.containerization.kubernetes.kubeconfig-support:
      revision: 0.1.0

spec:
  # TODO:(https://github.com/capactio/capact/issues/539): This will be moved as an optional input on Interface
  additionalInput:
    typeInstances:
      kubeconfig:
        typeRef:
          path: cap.type.containerization.kubernetes.kubeconfig
          revision: 0.1.0
        verbs: [ "get" ]

  outputTypeInstanceRelations:
    postgresql:
      uses:
        - psql-helm-release
        # TODO(https://github.com/capactio/capact/issues/537): renderer doesn't support relations for additionalInput.typeInstances
        #- kubeconfig

  requires:
    cap.core.type.platform:
      oneOf:
        - name: kubernetes
          revision: 0.1.0

  action:
    runnerInterface: argo.run
    args:
      workflow:
        entrypoint: postgres-install
        templates:
          - name: postgres-install
            inputs:
              artifacts:
                - name: kubeconfig
                  optional: true
            steps:
              - - name: helm-install
                  capact-action: helm.install
                  arguments:
                    artifacts:
                      # TODO(hack): here we cannot pass optional TI, see: https://github.com/capactio/capact/issues/538
                      # it works only because we test in Helm runner in file exists under given artifact path
                      # and we don't create relations to this kubeconfig.
                      - name: kubeconfig
                        from: "{{inputs.artifacts.kubeconfig}}"
                        optional: true
```

#### To all

via policy with `cap.*`

### To single step

via policy with `cap.interface.specific.name`

### Kubeconfig

1. Inject via `requires`
2. 

Cons:
- it's hidden from API
- support only one cluster per workflow?
- all runners need to be aware about kubeconfig

Issues:
- we need to be able to create relations (changes in engine)

### Multic-dev (argo)

It uses the kubeconfig. 

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: multi-cluster
  namespace: local
spec:
  entrypoint: main
  serviceAccountName: workflow
  templates:
    - name: main
      cluster: cluster-1            # easy to add during render process
      namespace: remote             # easy to add during render process
      container:
        image: docker/whalesay
```

Cons:
- how to create relations?? some changes in engine + info about target cluster, still via kubeconfig?

- Pros:
- second SA on target cluster
- runners don't know that they were executed on different cluster


## Questions
- Do we need to support multiple cluster installation in the same workflow? 
- 

### Virtual-kubelet-based approaches 

####

[Virtual Kubelet (VK)](https://github.com/virtual-kubelet/virtual-kubelet) is a “Kubernetes kubelet implementation that masquerades as a kubelet to connect Kubernetes to other APIs” [3].
Admiralty, Tensile-kube, and Liqo, adopt this approach. 


#### Admiralty

The Argo Workflows tutorial is out-dated. You can no longer enforce placement with the `multicluster.admiralty.io/clustername` annotation as they replaced that with a more idiomatic node selector instead.

- https://github.com/cwdsuzhou/super-scheduling based on https://github.com/virtual-kubelet/tensile-kube
- 

Cons:
- kubeconfig on cluster or user via CA
- needs to be installed on target cluster too
- don't have `CPU Requests  CPU Limits  Memory Requests  Memory Limits`
- uses old CRDs 
  - policy/v1beta1 PodDisruptionBudget is deprecated in v1.21+, unavailable in v1.25+; use policy/v1 PodDisruptionBudget
  - apiextensions.k8s.io/v1beta1 CustomResourceDefinition is deprecated in v1.16+, unavailable in v1.22+; use apiextensions.k8s.io/v1 CustomResourceDefinition
- Last update was in.. **TBD**
- It takes CPU + MEMORY **TBD**
- It's easy to schedule pod in all workloads cluster if you specify only `multicluster.admiralty.io/elect: ""` without `nodeSelector`. In our case it can be problematic.


- Pros:
- runners don't know that they were executed on different cluster


 
### Capact creates Capact

A dedicated Action (interface and implementation) that explicity accepts target, it can be any cluster, if not then create own. The easiest way and doens't require any additional new functionality.


diagram
1. Capact - "control plane for bootstrapping" one action is essentialy used, and the UI shows created TI (also can destroy/upgrade those cluster)
2. 

### Conclusion 

It's not solution for the long future, it's for now, later we can revisit that.  
