# 3. Install Crowd

Follow these steps to install Atlassian Crowd using the existing PostgreSQL installation.

1. Define required environmental variables:

   > **NOTE**: Use the PostgreSQL TypeInstance ID from the [Provision AWS RDS for PostgreSQL](./2-aws-rds-provisioning.md) tutorial.

   ```bash
   export CAPACT_DOMAIN_NAME={domain_name} # e.g. demo.cluster.capact.dev
   
   export POSTGRESQL_TI_ID=={typeinstance_id}
   ```

1. Create a file with input TypeInstances:

   ```bash
   cat > /tmp/crowd-ti.yaml << ENDOFFILE
   typeInstances:
     - name: "postgresql"
       id: "${POSTGRESQL_TI_ID}"
   ENDOFFILE
   ```   

1. Create a file with parameters:

   ```bash
   cat > /tmp/crowd-params.yaml << ENDOFFILE
   ingress:
     host: crowd.${CAPACT_DOMAIN_NAME}
   replicaCount: 1
   # use default values
   ENDOFFILE
   ```

   You can tweak the parameters based on the [input parameters JSON schema](https://github.com/capactio/capact/tree/main/och-content/type/productivity/crowd/install-input.yaml).

   > **NOTE:** It is recommended to start with a single Crowd replica before finishing the configuration process. Mind that the AWS EFS StorageClass is not compatible.

1. Create a dedicated Kubernetes Namespace:

   ```bash
   kubectl create namespace crowd
   ```

1. Create an Action:

   ```bash
   capact act create cap.interface.productivity.crowd.install --name crowd --namespace crowd --parameters-from-file /tmp/crowd-params.yaml --type-instances-from-file /tmp/crowd-ti.yaml
   ```

1. Wait until the Action is ready to run:

    ```bash
    capact act get crowd -n crowd
    ```   

1. Run the Action:

    ```bash
    capact act run crowd -n crowd
    ```

1. Watch the Action:

    ```bash
    capact act watch crowd -n crowd
    ```

1. Once the Action succeeded, list the output TypeInstances of the Action:

   ```bash
   capact act status crowd -n crowd
   ```

   Note the following values from output TypeInstances:
    - URL in **spec.value.host** property for **crowd-config** TypeInstance.
    - name (**spec.value.name**), and password (**spec.value.password**) of the **crowd-install-create-user-user** TypeInstance.
    - the database name (**spec.value.name**) of the **crowd-install-create-db-database** TypeInstance.

1. Get the PostgreSQL TypeInstance:

   > **NOTE:** This functionality is not yet available in the CLI. In this example, we will use `curl` to query the Capact Gateway GraphQL API directly.

   ```bash
   curl 'https://gateway.'"$CAPACT_DOMAIN_NAME"'/graphql' -H 'Accept-Encoding: gzip, deflate, br' -H 'Content-Type: application/json' -H 'Accept: application/json' -H 'Connection: keep-alive' -H 'DNT: 1' -H 'Origin: https://gateway.'$CAPACT_DOMAIN_NAME -H 'Authorization: Basic Z3JhcGhxbDp0MHBfczNjcjN0' --data-binary '{"query":"\nquery TypeInstance($typeInstance:ID!) {\n  typeInstance(id:$typeInstance) {\n    latestResourceVersion {\n      spec {\n        value\n      }\n    }\n  }\n}","variables":{"typeInstance":"'"$POSTGRESQL_TI_ID"'"}}' --compressed
   ```

   Note the database **host**.

1. Navigate to the Crowd URL noted in the previous step.
1. Click the **Set up Crowd** button.
1. Provide the Crowd license.

   > **NOTE:** You can generate a trial license using the "Atlassian Website" link during the "License" step.

1. Select **New installation** and click **Continue**.
1. As this is the most important part of Crowd configuration, watch the logs of Crowd server using this command in a new window to make sure the server behaves correctly:

   ```bash
   kubectl logs -n crowd -l=app.kubernetes.io/name=crowd -f
   ```

1. For the "Database configuration" step:
    - Select "JDBC connection".
    - In **Database** dropdown, pick "PostgreSQL".
    - In **JDBC URL** field, replace `localhost` with your database host noted from one of previous steps. Keep the `jdbc:postgresql://` prefix and port. Replace `crowd` suffix with your database name if it has different name.
    - In **Username**, fill database username noted from one of previous steps.
    - In **Password**, fill database password noted from one of previous steps.
    - Click **Continue**.

   ![Database configuration](./assets/crowd-db.png)

1. Wait for the Crowd server to respond. Observe the Crowd server logs to see if the configuration is processed.

   **Troubleshooting:** If you got the following responses:
    - "Connection timed out" error (or "This site can't be reached") or
    - "This database is not empty" error, and you're 100% sure the database is empty, Wait until the server processes the change, that is: it stops writing to the logs, and navigate to the root path of the website (`crowd.${CAPACT_DOMAIN_NAME}`). Click **Set up Crowd**, and you should be navigated to the configuration step after Database configuration.

1. Provide other parameters and finish Crowd configuration.
1. Crowd is ready to use.

**Next steps:** Navigate back to the [Introduction](./0-intro.md) and follow next steps.
