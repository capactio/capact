# 7. Install Rocket.chat

To deploy Rocket.chat on Kubernetes, follow the steps:

1. Export Capact cluster domain name as environment variable:

   ```bash
   export CAPACT_DOMAIN_NAME={domain_name} # e.g. demo.cluster.capact.dev
   ``` 

1. Create a file with installation parameters:

    ```bash
    cat > /tmp/rocketchat-params.yaml << ENDOFFILE
    host: rocketchat.${CAPACT_DOMAIN_NAME}
    replicaCount: 2
    resources:
      requests:
        memory: "2G"
        cpu: "1"
      limits:
        memory: "4G"
        cpu: "1"
    ENDOFFILE
    ```

1. Create a Kubernetes Namespace:

    ```bash
    kubectl create namespace rocketchat
    ```

1. Create an Action:
 
    ```bash
    capact action create cap.interface.productivity.rocketchat.install \
    --name rocketchat \
    --namespace rocketchat \
    --parameters-from-file /tmp/rocketchat-params.yaml
    ```

1. Wait for status `READY_TO_RUN`. Get status by running:

   ```bash
   capact --namespace rocketchat action get rocketchat
   ```

1. Run the Action:

   ```bash
   capact --namespace rocketchat action run rocketchat
   ```

1. Watch the Action:

   ```bash
   capact --namespace rocketchat action watch rocketchat
   ```

1. Once the Action is succeeded, view output TypeInstances:

   ```bash
   capact --namespace rocketchat action status rocketchat
   ```
    
🎉 Hooray! You successfully completed the tutorial. Be productive!
