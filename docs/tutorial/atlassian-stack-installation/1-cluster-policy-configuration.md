# 1. Configure Cluster Policy for AWS solutions

Configure preference for AWS solutions for Atlassian stack dependencies. Follow these steps:

1. Create new User in AWS dashboard [https://console.aws.amazon.com/iam/home?region=eu-west-1#/home](https://console.aws.amazon.com/iam/home?region=eu-west-1#/home)

   Add the following permissions:
    - **AmazonVPCFullAccess**
    - **AmazonRDSFullAccess**
    - **AmazonRDSDataFullAccess**
    - **IAMFullAccess**

   Note the access key and secret key.

1. Follow the tutorial ["Connect to Capact Gateway from local machine"](../eks-installation/README.md#connect-to-capact-gateway-from-local-machine) to be able to connect to Gateway.
1. Execute `helm get notes capact -n capact-system` on Bastion host and copy required headers.
1. Navigate to the `https://gateway.${CAPACT_DOMAIN_NAME}:8081` address in your web browser.
1. Create AWS Credentials TypeInstance:

    ```graphql
    mutation CreateTypeInstance {
      createTypeInstance(
        in: {
          typeRef: { path: "cap.type.aws.auth.credentials", revision: "0.1.0" }
          value: {
            accessKeyID: "{ACCESS_KEY}",
            secretAccessKey:"{SECRET_KEY}",
          }
          attributes: [
            { path: "cap.attribute.cloud.provider.aws", revision: "0.1.0" }
          ]
        }
      ) {
        id
        typeRef {
            path
            revision
          }
      }
    }
    ```

1. Export the ID of the newly created TypeInstance:

    ```bash
    export TI_ID={id}
    ```

1. Create a file with the new cluster policy:
   
    ```bash
    cat > /tmp/policy.yaml << ENDOFFILE
    apiVersion: 0.1.0
    rules:
       cap.interface.database.postgresql.install:
         oneOf:
           - implementationConstraints:
               attributes:
                 - path: "cap.attribute.cloud.provider.aws"
       cap.interface.aws.rds.postgresql.provision:
         oneOf:
           - implementationConstraints:
               attributes:
                 - path: "cap.attribute.cloud.provider.aws"
             injectTypeInstances:
               - id: ${TI_ID}
                 typeRef:
                   path: "cap.type.aws.auth.credentials"
                   revision: "0.1.0"
       cap.interface.analytics.elasticsearch.install:
         oneOf:
           - implementationConstraints:
               attributes:
                 - path: "cap.attribute.cloud.provider.aws"
       cap.interface.aws.elasticsearch.provision:
         oneOf:
           - implementationConstraints:
               attributes:
                 - path: "cap.attribute.cloud.provider.aws"
             injectTypeInstances:
               - id: ${TI_ID}
                 typeRef:
                   path: "cap.type.aws.auth.credentials"
                   revision: "0.1.0"
       cap.*:
         oneOf:
           - implementationConstraints:
               requires:
                 - path: "cap.core.type.platform.kubernetes"
           - implementationConstraints: {}
    ENDOFFILE
    ```

1. Update the cluster policy ConfigMap:

```bash
kubectl create configmap -n capact-system capact-engine-cluster-policy --from-file=cluster-policy.yaml=/tmp/policy.yaml -o yaml --dry-run=client | kubectl apply -f -
```

**Next steps:** Navigate back to the [main Readme](./README.md) and follow next steps.
