# Multi-cluster support

## Capact manifests definition

This section describes how the multi cluster can be described in Capact manifests. The idea is to use the already existing `requires` section and `additionalInput`.

When the Implementation requires `cap.core.type.platform.kubernetes`, the default `cap.core.type.platform.kubernetes` TypeInstance is used. This means that steps are executed on Kubernetes where Capact was installed. To override that you can use [Policy](https://capact.io/docs/next/feature/policies/overview) to inject other `cap.core.type.platform.kubernetes` TypeInstance which has the `kubeconfig`. This is information is used by our engine and steps are scheduled on target cluster with provided kubeconfig permissions.   

### Use cluster generated in umbrella workflow 

```yaml
kind: Implementation
metadata:
  displayName: "Boostrap PostgreSQL cluster"
  # ...
spec:
  additionalInput:
    typeInstances:
      kubernetes:
        typeRef:
          path: cap.core.type.platform.kubernets
          revision: 0.1.0
        verbs: [ "get" ]

  # No `requires` section - workflow doesn't depend on K8s directly.

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
                - name: kubernetes
                  optional: true
            steps:
              - - name: create-target-cluster
                  capact-when: kubernetes == nil
                  capact-action: k8s.deploy
              - - name: postgresql-install
                  capact-action: postgresql.install
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
                                - artifact: "{{steps.create-target-cluster.outputs.artifacts.kubernetes}}"
```

### Use pre-existing (registered) cluster

```yaml
kind: Implementation
metadata:
  # ...
  displayName: Install PostgreSQL database
spec:
  # ...
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
            steps:
              - - name: helm-install
                  capact-action: helm.install
```

If necessary we can introduce the `capact-target` property which can be used to mark a given step to be executed on a given cluster. Without `capact-target` step is not mutated and excuted where Argo was installed.
```yaml
# ...
  requires:
    cap.core.type.platform:
      oneOf:
        - name: kubernetes
          alias: k8s
          revision: 0.1.0

# ...
            steps:
              - - name: helm-install
                  capact-action: helm.install
                  capact-target: k8s # based on alias in requires section 
# ...
```

Capact user or admin can override the default Kubernetes cluster used in Implementation via Policy:

```yaml
rules:
  - interface: 
      path: "cap.interface.database.postgresql.install"
    oneOf:
      - implementationConstraints: # In first place, find and use an Implementation which:
          requires: # has the following Type references defined in the `spec.requires` property:
            - path: "cap.core.type.platform.kubernetes"
        inject:
          requiredTypeInstances: # For such Implementation, inject the following TypeInstance. 
            - id: 9038dcdc-e959-41c4-a690-d8ebf929ac0c
              description: "BigBang EU cluster"
```

### Consequences

To reduce the boilerplate and support multi-cluster in Capact, following items needs to be resolved:
1. Remove `cap.attribute.containerization.kubernetes.kubeconfig-support` from **attributes**.
2. Merge the `cap.type.containerization.kubernetes.kubeconfig` to `cap.core.type.platform.kubernets` as we don't have the [Type composition](https://capact.io/docs/feature/type-features#type-composition) yet.
3. Use Policy to inject other `cap.core.type.platform.kubernets`. We already use such approach for AWS and GCP credentials.
4. Add option to use `inject.requiredTypeInstances` via artifact name in Workflow Policy. This will solve https://github.com/capactio/capact/issues/538 as we will do that via Policy instead.
5. Solve [Support setting relations for optional TypeInstances in workflows](https://github.com/capactio/capact/issues/537) issue but also take into account TypeInstances from the `requires` section.

## Possible implementations

This section describes possible options on how to implement logic for syntax described in the [Capact manifests definition](#capact-manifests-definition) section.

### Kubernetes TypeInstance

Currently, we inject the `cap.attribute.containerization.kubernetes.kubeconfig-support`. This simply can be changed to `cap.type.containerization.kubernetes.kubeconfig` and  

Cons:
- support only one cluster per workflow
- all runners need to be aware about kubeconfig

Issues:
- we need to be able to create relations (changes in engine)

- k8s port-forward, helm runner impl

### Multi-cluster Workflows in Argo

Argo Workflows wants to support the [multi-cluster Workflows](https://github.com/argoproj/argo-workflows/issues/3523). There is a [draft PR](https://github.com/argoproj/argo-workflows/pull/6804) but without any information when this functionality will be merged.

The proposed solution adds `cluster` and `namespace` properties for container template.
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

This can be added during the render process by our [Argo Renderer](https://github.com/capactio/capact/tree/main/pkg/sdk/renderer/argo). Unfortunately, as this is still work in progress we cannot relay on it. It can be revisited in the future. 

**Pros:**
- Creates workload directly on target cluster, 
- No need to install agents on target cluster,
- Uses dedicated ServiceAccount to create and managed created resources on target cluster,
- Capact Runners don't know that they were executed on different cluster.

**Cons:**
- Logs cannot be accessed from the user interface.

### Virtual-kubelet-based approaches 

[Virtual Kubelet (VK)](https://github.com/virtual-kubelet/virtual-kubelet) is a “Kubernetes kubelet implementation that masquerades as a kubelet to connect Kubernetes to other APIs” [3].
Admiralty, Tensile-kube, and Liqo, adopt this approach. 

#### Admiralty

We have a dedicate Action to register an external cluster. Thanks to that we can install the
Action to prepare that cluster for multi-support ..

The Argo Workflows tutorial is out-dated. You can no longer enforce placement with the `multicluster.admiralty.io/clustername` annotation as they replaced that with a more idiomatic node selector instead.

**Cons:**
- needs to be installed on target cluster
- Last update was in.. **TBD**
- It takes CPU + MEMORY **TBD**
- It's easy to schedule pod in **all** registered target clusters if you specify only `multicluster.admiralty.io/elect: ""` without `nodeSelector`. In our case it can be problematic.
- Manifests in Helm chart don't have specified the CPU and memory requests and limits. This can be easily solved.
- Uses old Kubernetes manifest versions, which generates such warnings:
  - `policy/v1beta1 PodDisruptionBudget is deprecated in v1.21+, unavailable in v1.25+; use policy/v1 PodDisruptionBudget`
  - `apiextensions.k8s.io/v1beta1 CustomResourceDefinition is deprecated in v1.16+, unavailable in v1.22+; use apiextensions.k8s.io/v1 CustomResourceDefinition`

**Pros:**
- runners don't know that they were executed on different cluster
- as it's uses the proxy Pod concept, all functionality e.g. fetching logs, checking Pod status etc. can be executed on the main cluster. 

 
### Capact creates Capact

This clearly states that the Capact doesn't support the multi-cluster, but instead it's able to create a new Kubernetes clusters with Capact installed on it.   

In that approach, we have:
- "Capact control plane" for bootstrapping other Capact clusters. All created clusters are described via TypeInstance. This can be visible in UI where user admin can browser all provisioned Capact clusters and executed other Actions against those instances. For example, upgrade Capact cluster or destroy it. 
- "Capact" -

This is the easiest way and doesn't require any additional new functionality to be implemented or changed in Capact.

### Consequences

1. Create a dedicated Action (Interface and Implementation) that explicitly accepts target cluster. It can be any cluster, if not specified, workflow creates own. 

## Decision

**TBD**
