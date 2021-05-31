# 4. Install Bitbucket

Follow these steps to install [Atlassian Bitbucket Data Center](https://github.com/capactio/capact/tree/main/och-content/interface/productivity/bitbucket/install.yaml) on Kubernetes with the existing PostgreSQL instance.

1. Save your Bitbucket license in the `license.txt` file.

    >**NOTE:** You can generate a trial license from the [Atlassian Website](https://my.atlassian.com/license/evaluation).

1. Export the Capact cluster domain name, Bitbucket license and PostgreSQL TypeInstance ID as environment variables:

    >**NOTE**: Use the PostgreSQL TypeInstance ID from the [Provision AWS RDS for PostgreSQL](./2-aws-rds-provisioning.md) tutorial.

   ```bash
   export CAPACT_DOMAIN_NAME={domain_name} # e.g. demo.cluster.capact.dev
   export LICENSE_KEY_BASE64=$(/bin/cat license.txt | base64 )
   export POSTGRESQL_TI_ID={ti_id} 
   ``` 

1. Create a file with installation parameters:

    ```bash
    cat > /tmp/bb-params.yaml << ENDOFFILE
    # -- The Bitbucket replica count.
    # You can use more that one replica only when the licenseKeyInBase64 and sysadminCredentials are
    # specified. Otherwise, you need to configure Bitbucket after the initial startup and scale it up after
    # it is configured. 
    replicaCount: 2
    
    ingress:
      host: bitbucket.${CAPACT_DOMAIN_NAME}
    
    bitbucket:
      # -- The Bitbucket license key.
      # If specified, the license is automatically populated during Bitbucket setup.
      # Otherwise, it must be provided via browser after the initial startup.
      licenseKeyInBase64: ${LICENSE_KEY_BASE64}
   
      # -- The Bitbucket sysadmin credentials.
      # If specified, the credentials are automatically populated during Bitbucket setup.
      # Otherwise, it must be provided via browser after the initial startup.
      sysadminCredentials:
        displayName: "Capact Admin"
        emailAddress: admin@capact.io
        username: admin
        password: admin

      clustering:
        enabled: true
    
      resources:
        container:
          limits:
            cpu: "1"
            memory: "4G"
          requests:
            cpu: "1"
            memory: "2G"
   
    volumes:
      localHome:
        persistentVolumeClaim:
          create: true
      sharedHome:
        persistentVolumeClaim:
          create: true
          # Make sure that this matches your StorageClass name on your own K8s cluster.
          # It has to be a shared storage such as AWS EFS.
          storageClassName: "efs-sc"

    ENDOFFILE
    ```

2. Create a file with PostgreSQL TypeInstance ID:
 
    ```bash
    cat > /tmp/bb-ti.yaml << ENDOFFILE
    typeInstances:
      - name: "postgresql"
        id: "${POSTGRESQL_TI_ID}"
    ENDOFFILE
    ```

3. Create a Kubernetes Namespace:

    ```bash
    kubectl create namespace bitbucket
    ```

4. Create a Bitbucket Action:

    >**NOTE:** You must have a proper cluster policy configuration as described in the [Configure Cluster Policy to prefer AWS solutions](./1-cluster-policy-configuration.md) tutorial.
 
    ```bash
    capact action create cap.interface.productivity.bitbucket.install \
    --name bitbucket \
    --namespace bitbucket \
    --parameters-from-file /tmp/bb-params.yaml \
    --type-instances-from-file /tmp/bb-ti.yaml
    ```

5. Run the Action:

    ```bash
    capact action run bitbucket -n bitbucket
    ```

6. Watch the Action:

    ```bash
    capact action watch bitbucket -n bitbucket
    ```

7. Once the Action succeeded, list the output TypeInstances:

   ```bash
   capact action status bitbucket -n bitbucket
   ```

8. Find the **bitbucket-config** TypeInstance and copy-paste the **host** URL into your browser to open Bitbucket.
   
   >**NOTE:** To log in, use the credentials specified for the **sysadminCredentials** property in the input installation parameters. 

**Next steps:** Navigate back to the [Introduction](./0-intro.md) and follow next steps.
