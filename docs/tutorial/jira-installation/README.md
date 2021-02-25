# Jira installation

This tutorial shows the basic concepts of Voltron on Jira installation example.

### Introduction

The key benefit which Voltron brings is interchangeable dependencies. Cluster Admin may configure preferences for resolving the dependencies (e.g. to prefer cloud-based or on-premise solutions). As a result, the end-user is able to easily install applications with multiple dependencies without any knowledge of platform-specific configuration.

Apart from installing applications, Voltron makes it easy to:
- execute day-two operations (such as upgrade, backup, and restore)
- run any workflow (to process data, configure the system, run serverless workloads, etc. The possibilities are endless.)

Voltron aims to be a platform-agnostic solution. However, the very first Voltron implementation is based on Kubernetes.

### Goal

This instruction will guide you through the installation of Jira on the Kubernetes cluster using Voltron. 

Jira depends on the PostgreSQL database. Depending on the cluster configuration, with the Voltron project, you can install Jira with a managed Cloud SQL database or a locally deployed PostgreSQL Helm chart.

The bellow diagrams show possible scenarios:

**Install all Jira components in Kubernetes cluster**

![in-cluster](./assets/install-in-cluster.svg)

**Install Jira with external CloudSQL database**

![in-gcp](./assets/install-gcp.svg)

###  Prerequisites

* Install [`ocftool`](https://github.com/Project-Voltron/go-voltron/releases/tag/v0.1.0)
* Install [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* GKE cluster with fresh Voltron installation. See [installation tutorial](../voltron-installation/README.md). 
* Access to Google Cloud Platform for the scenario with CloudSQL.  

### Install all Jira components in Kubernetes cluster

By default, Voltron Engine has [cluster-policy](../../../deploy/kubernetes/charts/voltron/charts/engine/values.yaml) which prefers the Kubernetes solutions. 

```yaml
apiVersion: 0.1.0 # Defines syntax version for policy

rules: # Configures the following behavior for Engine during rendering Action
  cap.*:
    oneOf:
      - implementationConstraints: # prefer Implementation for Kubernetes
          requires:
            - path: "cap.core.type.platform.kubernetes"
              # any revision
      - implementationConstraints: {} # fallback to any Implementation
```

As a result all external solutions such as CloudSQL have lower priority and they are not selected. The below scenario shows how to install Jira with locally deploy PostgreSQL Helm chart.

#### Instruction

1. Create Kubernetes namespace:

	```bash
    export NAMESPACE=local-scenario
	kubectl create namespace $NAMESPACE
	```
 
1. Open GraphQL console:

    To obtain Gateway URL and authorization information, run:
    ```bash
    helm get notes -n voltron-system voltron    
    ```
   
   Based on returned response, navigate to GraphQL console and add required Authorization header.
   
   ![gql-auth](./assets/graphql-auth.png)

1. List all `cap.interface.productivity.*` Interfaces:

	<details><summary>Query</summary>

	```graphql
    query GetInterfaces {
      interfaces(filter: { pathPattern: "cap.interface.productivity.*" }) {
        path
        revisions {
          metadata {
            path
            displayName
          }
          revision
        }
      }
    }
	```

	</details>

	![list-interface](./assets/list-interface.png)

	The returned response on the right-hand side represents all available Actions that you can execute for **productivity** category. As you can see, you can install a Jira and Confluence instance. For manifests, there might be more revisions. Different revisions mean that the installation input/output parameters differ. It might be due to a new feature or one removed in a non-backward-compatible way.


1. Create an Action with the `cap.interface.productivity.jira.install` Interface:

	Before running the GraphQL mutation, you must add the `Namespace` header in the **HTTP HEADERS** section.

    <details><summary>Headers</summary>
	
	```json
	{
	 "Authorization": "....",
	 "Namespace": "local-scenario"
	}
	```
    
    Additionally, you must change the host value in input parameters.
 	
	<details><summary>Mutation</summary>

	```graphql
    mutation CreateAction {
      createAction(
        in: {
          name: "jira-instance"
          actionRef: { path: "cap.interface.productivity.jira.install" }
          input: {
            parameters: "{ \"host\": \"{REPLACE_WITH_HOST_NAME}\" }"
          }
        }
      ) {
        name
        createdAt
        renderedAction
        run
        status {
          phase
          timestamp
          message
          runner {
            status
          }
        }
      }
    }
	```

	</details>

	![create-action](./assets/create-action.png)

1. Get the status of the Action from the previous step:

	<details><summary>Query</summary>

	```graphql
	query GetAction {
	  action(name: "jira-instance") {
	    name
	    createdAt
	    renderedAction
	    run
	    status {
	      phase
	      timestamp
	      message
	      runner {
	        status
	      }
	    }
	  }
	}
	```

	</details>

	![get-action](./assets/get-action.png)

	In the previous step, when you created the Action, you saw the `INITIAL` phase  in the response. Now the Action is in `READY_TO_RUN`. It means that the Action was processed by the Engine and the Interface was resolved to a specific Implementation. As a user, you can verify that the rendered Action is what you expected. If the rendering is taking more time, you will see the `BEING_RENDERED` phase.

1. Run the rendered Action:

	In the previous step, the Action was in the `READY_TO_RUN` phase. It is not executed automatically, as the Engine waits for the user's approval. To execute it, you need to send such a mutation:

	<details><summary>Mutation</summary>

	```graphql
	mutation RunAction {
	  runAction(name: "jira-instance") {
	    name
	    createdAt
	    run
	  }
	}
	```

	</details>

	![run-action](./assets/run-action.png)

1. Check Action execution:
    
    **Using [Argo CLI](https://github.com/argoproj/argo-workflows/releases/tag/v2.12.8)**
    
    ```
    argo watch jira-instance -n $NAMESPACE
    ```
    
    **Using Argo UI**
    
    By default, the Argo UI doesn't have dedicated Ingress. You need to port-forward Service to your local machine: 
    
    ```
    kubectl -n argo port-forward svc/argo-server 2746
    ```
   
    Navigate to [http://localhost:2746](http://localhost:2746) to open Argo UI and check the currently running `jira-install` workflow.

1. Wait until the Action is in the `SUCCEEDED` phase:

	<details><summary>Query</summary>

	```graphql
	query GetAction {
	  action(name: "jira-instance") {
	    name
	    createdAt
	    renderedAction
	    run
	    status {
	      phase
	      timestamp
	      message
	      runner {
	        status
	      }
	    }
	  }
	}
	```

	</details>

1. Get Argo Workflow logs to check the uploaded TypeInstance ID: 
    
    **Using Argo CLI**
    
    ```bash
    argo logs jira-instance -n $NAMESPACE | grep -e 'upload_type_instances*'
    ```

    **Using Argo UI**
    
    From the **Workflows** view select `jira-instance`. Next, select the last step called `upload-output-type-instances-step` and get its logs. The logs contain the uploaded TypeInstance ID.

	![get-logs](./assets/get-logs.png)

1. Get the TypeInstance details: 

    Use the ID from the previous step and fetch the TypeInstance details.

	<details><summary>Query</summary>

	```graphql
    query GetTypeInstance {
      typeInstance(id: "{JIRA_CONFIG_ID}") {
        spec {
          value
          typeRef {
            path
          }
        }
      }
    }
	```

	</details>

	![get-type-instance](./assets/get-type-instance.png)

1. Open Jira console using the **host** value from the previous step:

    ![jira-installation](./assets/jira-installation.png)

1. When you are done, remove the Action and Helm charts:

    ```bash
    kubectl delete action jira-instance -n $NAMESPACE
    helm delete -n $NAMESPACE $(helm list -f="jira-software-*|postgresql-*" -q -n $NAMESPACE)
    ```

### Install Jira with external CloudSQL database

To change the Jira installation we need to adjust our cluster-policy to prefer the GCP solutions.  More information about policy configuration can be found [here](../../policy-configuration.md).

#### Instructions


1. Create the GCP Service Account JSON access key:
   
   	* Open [https://console.cloud.google.com](https://console.cloud.google.com) and select your project.
   
   	* On the left pane, go to **IAM & Admin** and select **Service accounts**.
   
   	* Click **Create service account**, name your account, and click **Create**.
   
   	* Set the `Cloud SQL Admin` role.
   
   	* Click **Create key** and choose `JSON` as a key type.
   
   	* Save the `JSON` file.
   
   	* Click **Done**.


1. Convert GCP Service Account to JavaScript format:

   ```bash
   cat {PATH_TO_GCP_SA_FILE} | sed -E 's/(^ *)"([^"]*)":/\1\2:/'
   ```

1. Create TypeInstance with GCP Service Account:

   	Before running the GraphQL mutation, you must replace the `value` parameter with output from the previous step. 
   
   ```graphql
    mutation CreateTypeInstance {
      createTypeInstance(
        in: {
          typeRef: { path: "cap.type.gcp.auth.service-account", revision: "0.1.0" }
          value: {} # Replace with SA in JS format
          attributes: [
            { path: "cap.attribute.cloud.provider.gcp", revision: "0.1.0" }
          ]
        }
      ) {
        metadata {
          id
        }
        spec {
          typeRef {
            path
            revision
          }
        }
      }
    }
   ```

1. Export TypeInstance UUID:

   In the response from the previous step, you have the TypeInstance ID, export it as environment variable:
   ```bash
   export TI_ID={TYPE_INSTANCE_ID}
   ```

1. Create a file with new cluster policy:

   ```yaml
   cat > /tmp/policy.yaml << ENDOFFILE
   apiVersion: 0.1.0
   rules:
     cap.interface.database.postgresql.install:
      oneOf:
        - implementationConstraints:
            attributes:
              - path: "cap.attribute.cloud.provider.gcp"
            requires:
              - path: "cap.type.gcp.auth.service-account"
          injectTypeInstances:
            - id: ${TI_ID}
              typeRef:
                path: "cap.type.gcp.auth.service-account"
                revision: "0.1.0"
     cap.*:
       oneOf:
         - implementationConstraints:
             requires:
               - path: "cap.core.type.platform.kubernetes"
         - implementationConstraints: {} # fallback to any Implementation
   ENDOFFILE
   ```
   >**NOTE**: Check [policy configuration document](../../policy-configuration.md) if you are not familiar with the above syntax.  

1. Update cluster policy ConfigMap:

   ```bash
   kubectl create configmap -n voltron-system voltron-engine-cluster-policy --from-file=cluster-policy.yaml=/tmp/policy.yaml -o yaml --dry-run=client | kubectl apply -f -
   ``` 

1. Create Kubernetes namespace:

	```bash
    export NAMESPACE=gcp-scenario
	kubectl create namespace $NAMESPACE
	```

1. Install Jira with new cluster policy:

   Cluster policy is updated to prefer the GCP solutions for PostgreSQL Interface. As a result, Engine during the render process will select a CloudSQL Implementation which is available in our OCH server.
   

   Repeat the steps from [Install all Jira components in Kubernetes cluster](#install-all-jira-components-in-kubernetes-cluster) in the `gcp-scenario` Namespace. Start with the 4th and remember to update Namespace value in the GraphQL *HTTP HEADERS** section.


1. When you are done, remove the Cloud SQL manually and delete Action:

    ```bash
    kubectl delete action jira-instance -n $NAMESPACE
    ```

### Behind the scenes

The following section extends the tutorial with additional topics, to learn Voltron concepts even deeper.

#### OCF manifests

A user consumes content stored in Open Capability Hub (OCH). The content is defined using Open Capability Format (OCF) manifests. OCF specification defines the shape of manifests that Voltron understands, such as Interface or Implementation.

To see all the manifest that OCH stores, navigate to the [OCH content structure](https://github.com/Project-Voltron/go-voltron/tree/master/och-content).

To see the Jira installation manifests, click on the following links:
 - [Jira installation Interface](https://github.com/Project-Voltron/go-voltron/tree/master/och-content/interface/productivity/jira/install.yaml) - a generic description of Jira installation (action name, input, and output - a concept similar to interfaces in programming languages),
 - [Jira installation Implementation](https://github.com/Project-Voltron/go-voltron/tree/master/och-content/implementation/atlassian/jira/install.yaml) - represents the dynamic workflow for Jira Installation.

#### Content development

To make it easier to develop new OCH content, we implemented a dedicated CLI. Currently, it exposes the validation feature for OCF manifests. It detects the manifest kind and OCF version to properly validate a given file. You can use it to validate one or multiple files at a single run.

To validate all OCH manifests, navigate to the repository root directory, and run the following command:
```
ocftool validate ./och-content/**/*.yaml
```

In the future, we will extend the `ocftool` with additional features, such as:
- manifests scaffolding,
- manifests submission,
- signing manifests.

###  Additional resources

If you want to learn more about the project check the [go-voltron](https://github.com/Project-Voltron/go-voltron) repository.

Here are some useful links:

- [Tutorial which shows the first steps on how to develop OCF content for Voltron.](../content-creation/README.md)
- The [OCF Draft v0.0.1](https://docs.google.com/document/d/1ud7xL3bXxEXtVPE8daA_DHYacKHMkn_jx6s7eaVT-NA/edit?usp=drive_web&ouid=115672498843496061020) document. 
- Documentation which contains various investigations, enhancement proposals, tutorials, Voltron architecture and development guidelines can be found [here](https://github.com/Project-Voltron/go-voltron/tree/master/docs),
- Google Drive folder with the [initial draft concepts](https://drive.google.com/drive/u/1/folders/1SBpIR0QUn9Rp68w6N3G-hqXdi1HfZQsn),
