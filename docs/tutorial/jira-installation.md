# Jira installation

This tutorial shows how to install Jira on Kubernetes cluster using Voltron project. 

In the provided solution steps are interchangeable, so you can install Jira with managed Cloud SQL database or locally deployed PostgreSQL Helm chart.

The bellow diagrams show possible scenarios:

**Install all Jira components in Kubernetes cluster**

![in-cluster](./assets/install-in-cluster.svg)

**Install Jira with external CloudSQL database**

![in-gcp](./assets/install-gcp.svg)

###  Prerequisites

* Install Docker
* Install [`kind`](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
* Install [`ocftool`](https://github.com/Project-Voltron/go-voltron/releases/tag/v0.0.1-alpha.1)
* Install [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* Access to Google Cloud Platform 
	
### Install Jira with managed Cloud SQL

The below scenario installs Jira with Cloud SQL database because Engine detected the `gcp-credentials` secret in `gcp-scenario` namespace. 

1. Clone the `go-voltron` repository in the `v0.0.1-alpha.1` version:

	```bash
	git clone --depth 1 --branch release-0.1 https://github.com/Project-Voltron/go-voltron.git
	cd ./go-voltron
	```

1. Create a local cluster with Voltron components installed:

	```bash
	MOCK_OCH_GRAPHQL=true make dev-cluster
	```

	> **NOTE:** This takes around 5 minutes to finish.

1. Create Kubernetes namespace:

	```bash
	kubectl create namespace gcp-scenario
	```

1. Create the Kubernetes Secret which contains a GCP JSON access key:
   
   	* Open [https://console.cloud.google.com](https://console.cloud.google.com) and select your project.
   
   	* On the left pane, go to **Identity** and select **Service accounts**.
   
   	* Click **Create service account**, name your account, and click **Create**.
   
   	* Set the `Cloud SQL Admin` role.
   
   	* Click **Create key** and choose `JSON` as a key type.
   
   	* Save the `JSON` file.
   
   	* Create a Secret from the JSON file by running this command:
   
   		```bash
   		kubectl create secret generic gcp-credentials --from-file=sa.json={FILENAME}  --namespace gcp-scenario
   		```
   
   	* Click **Done**.

1. Navigate to [https://gateway.voltron.local](https://gateway.voltron.local) to open the GraphQL console.
 
1. Add the following Authorization header:

	```json
	{
	  "Authorization": "Basic Z3JhcGhxbDp0MHBfczNjcjN0"
	}
	```

	![gql-auth](./assets/graphql-auth.png)

1. List all available Interfaces:

	<details><summary>Query</summary>

	```graphql
	query GetInterfaces {
	  interfaces {
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

	The returned response on the right-hand side represents all available Actions that you can execute. As you can see, you can install and upgrade a Jira instance. For installation, there are two revisions. Different revisions mean that the installation input/output parameters differ. It might be due to a new feature or one removed in a non-backward-compatible way.

1. Create an Action with the `cap.interface.productivity.jira.install` Interface:

	Before running the GraphQL mutation, you must add the `Namespace` header in the **HTTP HEADERS** section.

    <details><summary>Headers</summary>
	
	```json
	{
	 "Authorization": "Basic Z3JhcGhxbDp0MHBfczNjcjN0",
	 "Namespace": "gcp-scenario"
	}
	```
    
    </details>

	<details><summary>Mutation</summary>

	```graphql
	mutation CreateAction {
	  createAction(in: { name: "jira-instance", action: "cap.interface.productivity.jira.install" }) {
	    name
	    createdAt
	    path
	    renderedAction
	    run
	    status {
	      condition
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
	    path
	    renderedAction
	    run
	    status {
	      condition
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

	In the previous step, when you created the Action, you saw the `INITIAL` phase  in the response. Now the Action is in `READY_TO_RUN`. It means that the Action was processed by the Engine and the Interface was resolved to a specific Implementation. As a user, you can verify that the rendered Action is what you expected. If the rendering is taking more time, you will see the `RENDERING` phase.

1. Run the rendered Action:

	In the previous step, the Action was in the `READY_TO_RUN` phase. It is not executed automatically, as the Engine waits for the user's approval. To execute it, you need to send such a mutation:

	<details><summary>Mutation</summary>

	```graphql
	mutation RunAction {
	  runAction(name: "jira-instance") {
	    name
	    createdAt
	    path
	    run
	  }
	}
	```

	</details>

	![run-action](./assets/run-action.png)

1. Navigate to [https://argo.voltron.local](https://gateway.voltron.local) to open Argo UI and check the currently running `jira-install` workflow.

1. Navigate to your [GCP CloudSQL Console](https://console.cloud.google.com/sql/instances). You should see that a new CloudSQL instance is being created.
    
    ![gcp-cloudsql](./assets/gcp-cloudsql.png)
    
1.	Wait until the Action is in the `SUCCEEDED` phase:

	<details><summary>Query</summary>

	```graphql
	query GetAction {
	  action(name: "jira-instance") {
	    name
	    createdAt
	    path
	    renderedAction
	    run
	    status {
	      condition
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

1.	Get Argo Workflow logs to check the uploaded TypeInstance ID: 

    From the **Workflows** view select `jira-instance`. Next, select the last step called `upload-type-instances` and get its logs. The logs contain the uploaded TypeInstance ID.

	![get-logs](./assets/get-logs.png)

1.	Get the TypeInstance details: 

    Use the ID from the previous step and fetch the TypeInstance details.

	<details><summary>Query</summary>

	```graphql
    query GetTypeInstance {
      typeInstance(id: "13343627-ab4a-4fbf-a312-f96f11b07d0b") {
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

1.	Open Jira console using the **hostname** value from the previous step:

	The installed Jira URL is: [https://jira-cloud.cluster.local/](https://gateway.voltron.local).

### Install Jira with on-premise PostgreSQL database

Repeat the steps from [Install Jira with managed Cloud SQL](#install-jira-with-managed-cloud-sql) in a different namespace and skip the 4th and 9th step. If the `gcp-credentials` secret does not exist, Engine renders the Workflow with the step to locally deploy PostgreSQL Helm chart. Shortly, we will provide dedicated policies that will allow cluster admins to set more complicated rules e.g. to associate GCP subscription with a given user.

###  Additional resources

If you want to learn more about the project check the [go-voltron](https://github.com/Project-Voltron/go-voltron) repository.

Here are some useful links:

- OCH content showed during the demo:
    * [OCH content structure](https://github.com/Project-Voltron/go-voltron/tree/master/och-content)
    * [Jira installation Interface](https://github.com/Project-Voltron/go-voltron/tree/master/och-content/interface/productivity/jira/install.yaml)
    * [Jira installation Implementation](https://github.com/Project-Voltron/go-voltron/tree/master/och-content/implementation/atlassian/jira/install.yaml)

- Documentation which contains various investigations, enhancement proposals, tutorials, Voltron architecture and development guidelines can be found [here](https://github.com/Project-Voltron/go-voltron/tree/master/docs),

- Google Drive folder with the [initial draft concepts](https://drive.google.com/drive/u/1/folders/1SBpIR0QUn9Rp68w6N3G-hqXdi1HfZQsn),

- The [OCF Draft v0.0.1](https://docs.google.com/document/d/1ud7xL3bXxEXtVPE8daA_DHYacKHMkn_jx6s7eaVT-NA/edit?usp=drive_web&ouid=115672498843496061020) document. 
